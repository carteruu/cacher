package cacher

import (
	"context"
	"errors"
	"golang.org/x/sync/singleflight"
	"math/rand"
	"reflect"
	"strconv"
	"time"
)

type (
	// Cacher 缓存
	Cacher struct {
		repo     Repo                       //
		expire   time.Duration              //缓存保留时长
		sf       singleflight.Group         //
		typeConv map[typePair]TypeConverter //
	}
	// Repo 存储库接口，通过实现该接口，可以支持不同类型的存储方式
	Repo interface {
		// Get 获取
		//缓存不存在时，需要返回 nil,nil
		Get(ctx context.Context, key string) (interface{}, error)
		// Set 保存
		Set(ctx context.Context, key string, value interface{}, expire time.Duration) error
		// Del 删除
		Del(ctx context.Context, key string) error
	}
	// TypeConverter 类型转换器
	TypeConverter struct {
		SrcType interface{}
		DstType interface{}
		Fn      func(src interface{}) (interface{}, error)
	}
	Option struct {
		Expire         time.Duration   //缓存保留时长
		NilData        interface{}     //空缓存数据
		NilCacheExpire time.Duration   //空缓存保留时长。小于等于0时，不保存空缓存
		Converters     []TypeConverter //转换器
	}
	typePair struct {
		DstType reflect.Type
		SrcType reflect.Type
	}
)

var (
	//默认转换器
	typeConverters = []TypeConverter{
		{
			SrcType: "",
			DstType: false,
			Fn: func(src interface{}) (interface{}, error) {
				return strconv.ParseBool(src.(string))
			},
		}, {
			SrcType: "",
			DstType: 0,
			Fn: func(src interface{}) (interface{}, error) {
				return strconv.Atoi(src.(string))
			},
		}, {
			SrcType: "",
			DstType: uint(0),
			Fn: func(src interface{}) (interface{}, error) {
				val, err := strconv.ParseUint(src.(string), 10, 64)
				if err != nil {
					return nil, err
				}
				return uint(val), nil
			},
		}, {
			SrcType: "",
			DstType: float64(0),
			Fn: func(src interface{}) (interface{}, error) {
				return strconv.ParseFloat(src.(string), 64)
			},
		},
		{
			SrcType: []byte{},
			DstType: false,
			Fn: func(src interface{}) (interface{}, error) {
				return strconv.ParseBool(string(src.([]byte)))
			},
		}, {
			SrcType: []byte{},
			DstType: 0,
			Fn: func(src interface{}) (interface{}, error) {
				return strconv.Atoi(string(src.([]byte)))
			},
		}, {
			SrcType: []byte{},
			DstType: uint(0),
			Fn: func(src interface{}) (interface{}, error) {
				val, err := strconv.ParseUint(string(src.([]byte)), 10, 64)
				if err != nil {
					return nil, err
				}
				return uint(val), nil
			},
		}, {
			SrcType: []byte{},
			DstType: float64(0),
			Fn: func(src interface{}) (interface{}, error) {
				return strconv.ParseFloat(string(src.([]byte)), 64)
			},
		},
	}
)

func New(repo Repo, expire time.Duration) *Cacher {
	if expire <= 0 {
		panic(errors.New("缓存保存时长 expire 必须大于0"))
	}
	cache := Cacher{
		repo:     repo,
		expire:   expire,
		sf:       singleflight.Group{},
		typeConv: make(map[typePair]TypeConverter, len(typeConverters)),
	}
	for _, conv := range typeConverters {
		if err := cache.RegisterConverter(conv); err != nil {
			panic(err)
		}
	}
	return &cache
}

// RegisterConverter 注册类型转换器
func (c *Cacher) RegisterConverter(converter TypeConverter) error {
	if converter.SrcType == nil || converter.DstType == nil || converter.Fn == nil {
		return errors.New("转换器错误")
	}
	c.typeConv[typePair{SrcType: reflect.TypeOf(converter.SrcType), DstType: reflect.TypeOf(converter.DstType)}] = converter
	return nil
}

// Get 缓存存在时，返回缓存数据；不存在时，并发请求只有一个能到数据库查询，其他的等待。多个 goroutine 共享查询结果
//返回值：是否命中缓存，空缓存也返回 true
func (c *Cacher) Get(
	ctx context.Context,
	key string, //缓存键
	queryFn func() (interface{}, error),
	v interface{},
) (bool, error) {
	return c.GetWithOption(ctx, key, queryFn, v, nil)
}

func (c *Cacher) GetWithOption(
	ctx context.Context,
	key string,
	queryFunc func() (interface{}, error),
	v interface{},
	optFn func(opt *Option)) (useCache bool, _ error) {
	if key == "" {
		return false, errors.New("缓存键 key 不能为空字符串")
	}
	if queryFunc == nil {
		return false, errors.New("查询方法 queryFunc 不能为空")
	}

	opt := Option{Expire: c.expire}
	if optFn != nil {
		optFn(&opt)
	}
	if err := opt.Valid(); err != nil {
		return false, err
	}

	to := indirect(reflect.ValueOf(v))
	toType, _ := indirectType(to.Type())

	if toType.Kind() == reflect.Interface {
		toType, _ = indirectType(reflect.TypeOf(to.Interface()))
		oldTo := to
		to = reflect.New(reflect.TypeOf(to.Interface())).Elem()
		defer func() {
			oldTo.Set(to)
		}()
	}

	//查询缓存
	cacheData, err := c.repo.Get(ctx, key)
	//查询缓存错误
	if err != nil {
		return false, err
	}
	from := reflect.ValueOf(cacheData)
	useCache = true
	if !from.IsValid() {
		//没有缓存
		sfVal, err, _ := c.sf.Do(key, func() (interface{}, error) {
			//调用传入的查询数据的方法，查询数据
			queryData, err := queryFunc()
			if err != nil {
				return nil, err
			}
			//查询数据为空
			if queryData == nil {
				//设置空缓存
				if !opt.isCacheNil() {
					return nil, nil
				}
				nilFrom := reflect.ValueOf(opt.NilData)
				if !nilFrom.IsValid() {
					nilFrom = reflect.Zero(toType)
				}
				if err := c.repo.Set(ctx, key, nilFrom.Interface(), opt.NilCacheExpire); err != nil {
					return nil, err
				}
				return nilFrom.Interface(), nil
			}
			//设置缓存
			//缓存时长,加一个小于 十分之一缓存时间 的随机数，避免缓存雪崩
			cacheExpire := opt.Expire + time.Duration(rand.Int63n(int64(opt.Expire)/10))
			if err := c.repo.Set(ctx, key, queryData, cacheExpire); err != nil {
				return nil, err
			}
			return queryData, nil
		})
		if err != nil {
			return false, err
		}
		if sfVal == nil {
			return false, nil
		}
		from = reflect.ValueOf(sfVal)
		useCache = false
	}
	//先使用option的转换器
	fromType, _ := indirectType(from.Type())
	for _, conv := range opt.Converters {
		if fromType == reflect.TypeOf(conv.SrcType) && toType == reflect.TypeOf(conv.DstType) {
			val, err := conv.Fn(from.Interface())
			if err != nil {
				return false, err
			}
			if val != nil {
				to.Set(reflect.ValueOf(val))
			} else {
				to.Set(reflect.Zero(to.Type()))
			}
			return useCache, nil
		}
	}
	//再尝试类型转换
	if from.CanConvert(toType) {
		to.Set(from.Convert(toType))
		return useCache, nil
	}
	//最后尝试注册的类型转换器
	if conv, ok := c.typeConv[typePair{SrcType: fromType, DstType: toType}]; ok {
		val, err := conv.Fn(from.Interface())
		if err != nil {
			return false, err
		}
		if val != nil {
			to.Set(reflect.ValueOf(val))
		} else {
			to.Set(reflect.Zero(to.Type()))
		}
		return useCache, nil
	}
	return false, errors.New("不支持的类型转换")
}

// Del 删除缓存
func (c *Cacher) Del(ctx context.Context, key string) error {
	return c.repo.Del(ctx, key)
}

func (o Option) Valid() error {
	if o.Expire <= 0 {
		return errors.New("expire need bigger 0")
	}
	return nil
}

//是否保存空缓存
func (o Option) isCacheNil() bool {
	return o.NilCacheExpire > 0
}

func indirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func indirectType(reflectType reflect.Type) (_ reflect.Type, isPtr bool) {
	for reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
		isPtr = true
	}
	return reflectType, isPtr
}
