package cacher_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/carteruu/cacher"
	"reflect"
	"testing"
	"time"
)

var (
	repoBytesGetMap = map[string]repoGetMapVal{
		"string":      {data: []byte("rrte432423"), err: nil},
		"-int":        {data: []byte("-54123"), err: nil},
		"int":         {data: []byte("1423432"), err: nil},
		"uint:0":      {data: []byte("0"), err: nil},
		"uint":        {data: []byte("123"), err: nil},
		"float64":     {data: []byte("1.1234"), err: nil},
		"person-1":    {data: personObjBs, err: nil},
		"personArr":   {data: personArrBs, err: nil},
		"personSlice": {data: personSliceBs, err: nil},
		"personMap":   {data: personMapBs, err: nil},
		"nil":         {data: nil, err: nil},
	}
)

//需要测试的用例：
//查询数据类型：int、uint、float、字符串、结构体，包含任意元素的数组、切片、map
//缓存数据类型：字节切片、字符串、接口（原数据类型）
//查询数据状态：非空、空、异常、没有数据错误（NeedCacheNil）
//缓存数据状态：非空、空、错误、空缓存错误
//没有数据错误（NeedCacheNil）时，是否需要设置空缓存
func TestCache_Singleflight_Bytes(t *testing.T) {
	type fields struct {
		repo cacher.Repo
	}
	type args struct {
		key       string
		queryFunc func() (interface{}, error)
		expire    time.Duration
		v         interface{}
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantUseCache bool
		wantErr      error
		wantData     interface{}
	}{
		//缓存数据类型：字节切片
		{
			name: "查询：字节切片",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "string",
				queryFunc: func() (interface{}, error) {
					return nil, notNeedCall
				},
				expire: time.Second * 10,
				v:      &[]byte{},
			},
			wantUseCache: true,
			wantErr:      nil,
			wantData:     []byte("rrte432423"),
		}, {
			name: "查询：字节切片，缓存：不存在",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "nil",
				queryFunc: func() (interface{}, error) {
					return []byte("rrte42232423"), nil
				},
				expire: time.Second * 10,
				v:      &[]byte{},
			},
			wantUseCache: false,
			wantErr:      nil,
			wantData:     []byte("rrte42232423"),
		}, {
			name: "缓存：不存在，查询：不存在，不保存空缓存",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "nil",
				queryFunc: func() (interface{}, error) {
					return nil, nil
				},
				expire: time.Second * 10,
				v:      &[]byte{},
			},
			wantUseCache: false,
			wantErr:      nil,
			wantData:     []byte{},
		}, {
			name: "查询：int",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "int",
				queryFunc: func() (interface{}, error) {
					return nil, notNeedCall
				},
				expire: time.Second * 10,
				v:      &vInt,
			},
			wantUseCache: true,
			wantErr:      nil,
			wantData:     1423432,
		}, {
			name: "查询：int 负数",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "-int",
				queryFunc: func() (interface{}, error) {
					return nil, notNeedCall
				},
				expire: time.Second * 10,
				v:      &vInt,
			},
			wantUseCache: true,
			wantErr:      nil,
			wantData:     -54123,
		}, {
			name: "查询：uint:0",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "uint:0",
				queryFunc: func() (interface{}, error) {
					return nil, notNeedCall
				},
				expire: time.Second * 10,
				v:      &vUint,
			},
			wantUseCache: true,
			wantErr:      nil,
			wantData:     uint(0),
		}, {
			name: "查询：uint",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "uint",
				queryFunc: func() (interface{}, error) {
					return nil, notNeedCall
				},
				expire: time.Second * 10,
				v:      &vUint,
			},
			wantUseCache: true,
			wantErr:      nil,
			wantData:     uint(123),
		}, {
			name: "查询：float64",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "float64",
				queryFunc: func() (interface{}, error) {
					return nil, notNeedCall
				},
				expire: time.Second * 10,
				v:      &vFloat64,
			},
			wantUseCache: true,
			wantErr:      nil,
			wantData:     1.1234,
		}, {
			name: "查询：字符串",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "string",
				queryFunc: func() (interface{}, error) {
					return nil, notNeedCall
				},
				expire: time.Second * 10,
				v:      &vString,
			},
			wantUseCache: true,
			wantErr:      nil,
			wantData:     "rrte432423",
		}, {
			name: "查询：字符串，缓存：不存在",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "nil",
				queryFunc: func() (interface{}, error) {
					return "rrte42232423", nil
				},
				expire: time.Second * 10,
				v:      &vString,
			},
			wantUseCache: false,
			wantErr:      nil,
			wantData:     "rrte42232423",
		}, {
			name: "查询：结构体",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "person-1",
				queryFunc: func() (interface{}, error) {
					return nil, notNeedCall
				},
				expire: time.Second * 10,
				v:      &vPerson,
			},
			wantUseCache: true,
			wantErr:      nil,
			wantData:     personObj,
		}, {
			name: "查询：结构体，缓存：不存在",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "nil",
				queryFunc: func() (interface{}, error) {
					return personObj1, nil
				},
				expire: time.Second * 10,
				v:      &vPerson,
			},
			wantUseCache: false,
			wantErr:      nil,
			wantData:     personObj1,
		},
		{
			name: "查询：结构体数组",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "personArr",
				queryFunc: func() (interface{}, error) {
					return nil, notNeedCall
				},
				expire: time.Second * 10,
				v:      &vPersonArr,
			},
			wantUseCache: true,
			wantErr:      nil,
			wantData:     personArr,
		}, {
			name: "查询：结构体数组，缓存：不存在",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "nil",
				queryFunc: func() (interface{}, error) {
					return personArr1, nil
				},
				expire: time.Second * 10,
				v:      &vPersonArr,
			},
			wantUseCache: false,
			wantErr:      nil,
			wantData:     personArr1,
		}, {
			name: "查询：结构体数组，缓存没有数据，查询也没有数据，希望原样返回",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "nil",
				queryFunc: func() (interface{}, error) {
					return nil, nil
				},
				expire: time.Second * 10,
				v:      &vPersonArr1,
			},
			wantUseCache: false,
			wantErr:      nil,
			wantData:     [2]person{},
		},
		{
			name: "查询：结构体切片",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "personSlice",
				queryFunc: func() (interface{}, error) {
					return nil, notNeedCall
				},
				expire: time.Second * 10,
				v:      &vPersonSlice,
			},
			wantUseCache: true,
			wantErr:      nil,
			wantData:     personSlice,
		}, {
			name: "查询：结构体切片，缓存：不存在",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "nil",
				queryFunc: func() (interface{}, error) {
					return personSlice1, nil
				},
				expire: time.Second * 10,
				v:      &vPersonSlice1,
			},
			wantUseCache: false,
			wantErr:      nil,
			wantData:     personSlice1,
		}, {
			name: "查询：结构体切片，缓存没有数据，查询也没有数据，希望原样返回",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "nil",
				queryFunc: func() (interface{}, error) {
					return nil, nil
				},
				expire: time.Second * 10,
				v:      &vPersonSlice2,
			},
			wantUseCache: false,
			wantErr:      nil,
			wantData:     []person(nil),
		},
		{
			name: "查询：map[string]结构体",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "personMap",
				queryFunc: func() (interface{}, error) {
					return nil, notNeedCall
				},
				expire: time.Second * 10,
				v:      &vPersonMap,
			},
			wantUseCache: true,
			wantErr:      nil,
			wantData:     personMap,
		}, {
			name: "查询：map[string]结构体，缓存：不存在",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "nil",
				queryFunc: func() (interface{}, error) {
					return personMap1, nil
				},
				expire: time.Second * 10,
				v:      &vPersonMap1,
			},
			wantUseCache: false,
			wantErr:      nil,
			wantData:     personMap1,
		}, {
			name: "查询：map[string]结构体，缓存没有数据，查询也没有数据，希望原样返回",
			fields: fields{
				repo: &repoBytes{},
			},
			args: args{
				key: "nil",
				queryFunc: func() (interface{}, error) {
					return nil, nil
				},
				expire: time.Second * 10,
				v:      &vPersonMap2,
			},
			wantUseCache: false,
			wantErr:      nil,
			wantData:     map[string]person(nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cacher.New(tt.fields.repo, 10*time.Second)
			c.RegisterConverter(cacher.TypeConverter{
				SrcType: []byte{},
				DstType: person{},
				Fn: func(src interface{}) (interface{}, error) {
					var p person
					err := json.Unmarshal(src.([]byte), &p)
					if err != nil {
						return nil, err
					}
					return p, nil
				},
			})
			c.RegisterConverter(cacher.TypeConverter{
				SrcType: []byte{},
				DstType: []person{},
				Fn: func(src interface{}) (interface{}, error) {
					var p []person
					if err := json.Unmarshal(src.([]byte), &p); err != nil {
						return nil, err
					}
					return p, nil
				},
			})
			c.RegisterConverter(cacher.TypeConverter{
				SrcType: []byte{},
				DstType: map[string]person{},
				Fn: func(src interface{}) (interface{}, error) {
					var p map[string]person
					err := json.Unmarshal(src.([]byte), &p)
					if err != nil {
						return nil, err
					}
					return p, nil
				},
			})
			c.RegisterConverter(cacher.TypeConverter{
				SrcType: []byte{},
				DstType: [2]person{},
				Fn: func(src interface{}) (interface{}, error) {
					var p [2]person
					err := json.Unmarshal(src.([]byte), &p)
					if err != nil {
						return nil, err
					}
					return p, nil
				},
			})
			useCache, err := c.Get(context.Background(), tt.args.key, tt.args.queryFunc, tt.args.v)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Singleflight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if useCache != tt.wantUseCache {
				t.Errorf("Singleflight() useCache = %v, wantUseCache %v", useCache, tt.wantUseCache)
				return
			}
			//v是引用，需要解引用
			if !reflect.DeepEqual(tt.wantData, reflect.ValueOf(tt.args.v).Elem().Interface()) {
				t.Errorf("Singleflight() v = %v, wantData %v", reflect.ValueOf(tt.args.v).Elem().Interface(), tt.wantData)
				return
			}
		})
	}
}

type repoBytes struct{}

func (r *repoBytes) Del(ctx context.Context, keys ...string) error {
	panic("implement me")
}

func (r *repoBytes) Get(ctx context.Context, key string) (interface{}, error) {
	if val, ok := repoBytesGetMap[key]; ok {
		return val.data, val.err
	}
	return nil, fmt.Errorf("不支持的key")
}

func (r *repoBytes) Set(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	return nil
}
