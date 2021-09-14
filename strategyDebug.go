package main

import (
	"container/list"
	"fmt"
)

func strategyDebug(cache list.List) (bool, bool, uint64, error) {
	for e := cache.Back(); e != nil; e = e.Prev() {
		fmt.Println(e.Value) // print out the elements
	}
	return true, false, 12, nil
}
