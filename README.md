# cacher

## Usage

```go 
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/carteruu/cacher"
)

func main() {
	cntFn()
	cntFn() //use cache

	personFn()
	personFn() //use cache

	cacheNilFn()
	cacheNilFn() //use nil cache
}

type localRepo struct {
	cache map[string]interface{}
}

func (r *localRepo) Get(_ context.Context, key string) (interface{}, error) {
	data, exist := r.cache[key]
	if !exist {
		return nil, nil
	}
	return data, nil
}

func (r *localRepo) Set(_ context.Context, key string, value interface{}, expire time.Duration) error {
	r.cache[key] = value
	return nil
}

func (r *localRepo) Del(_ context.Context, key string) error {
	delete(r.cache, key)
	return nil
}

var cache = cacher.New(&localRepo{cache: make(map[string]interface{})}, 10*time.Second)

type person struct {
	name string
	age  int
}

func cntFn() {
	var cnt int
	useCache, err := cache.Get(
		context.Background(),
		"cnt-key",
		func() (interface{}, error) {
			fmt.Printf("query data cnt\n")
			return 99, nil
		},
		&cnt,
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("use cache=%v, cnt=%v\n", useCache, cnt)
}

func personFn() {
	var p person
	useCache, err := cache.Get(
		context.Background(),
		"p-key",
		func() (interface{}, error) {
			fmt.Printf("query data person\n")
			return person{
				name: "aa",
				age:  8,
			}, nil
		},
		&p,
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("use cache=%v, p=%+v\n", useCache, p)
}

func cacheNilFn() {
	var p person
	useCache, err := cache.GetWithOption(
		context.Background(),
		"p-n-key",
		func() (interface{}, error) {
			fmt.Printf("query data person: nil\n")
			return nil, nil
		},
		&p,
		func(opt *cacher.Option) {
			opt.NilCacheExpire = 1 * time.Second
		},
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("use cache=%v, p=%+v\n", useCache, p)
}

```