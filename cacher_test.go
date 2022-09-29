package cacher_test

import (
	"encoding/json"
	"fmt"
)

var (
	notNeedCall   = fmt.Errorf("不应该调用查询数据的方法")
	vString       string
	vInt          int
	vUint         uint
	vFloat64      float64
	vPerson       person
	vPersonArr    [2]person
	vPersonArr1   [2]person
	vPersonSlice  []person
	vPersonSlice1 []person
	vPersonSlice2 []person
	vPersonMap    map[string]person
	vPersonMap1   map[string]person
	vPersonMap2   map[string]person

	personObj = person{
		Name: "name-1",
		Age:  11,
		Address: address{
			Province: "广东",
			City:     "广州",
		},
	}
	personObj1 = person{
		Name: "name-2",
		Age:  22,
		Address: address{
			Province: "广东",
			City:     "广州",
		},
	}
	personArr    = [2]person{personObj, personObj1}
	personArr1   = [2]person{personObj1, personObj}
	personSlice  = []person{personObj, personObj1}
	personSlice1 = []person{personObj1, personObj}
	personMap    = map[string]person{"personObj": personObj, "personObj1": personObj1}
	personMap1   = map[string]person{"1": personObj1, "0": personObj}

	personObjBs, _   = json.Marshal(personObj)
	personArrBs, _   = json.Marshal(personArr)
	personSliceBs, _ = json.Marshal(personSlice)
	personMapBs, _   = json.Marshal(personMap)
)

type (
	repoGetMapVal struct {
		data interface{}
		err  error
	}
	address struct {
		Province string `json:"province"`
		City     string `json:"city"`
	}
	person struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Address address `json:"address"`
	}
)
