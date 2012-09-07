package main

import (
	"reflect"
)

type compareFunc func(a interface{}, b interface{}) bool

func swap(a, b reflect.Value) {
	temp := reflect.Indirect(reflect.New(a.Type()))
	temp.Set(a)
	a.Set(b)
	b.Set(temp)
}

func sortSlice(a interface{}, cmp compareFunc) { // a should be a slice
	v := reflect.ValueOf(a)
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		return
	}
	var qsort func(int, int)
	qsort = func(l, r int) {
		if l > r {
			return
		}
		i := l
		j := (r-l)/2 + l
		swap(v.Index(i), v.Index(j))
		j = l
		for i = l + 1; i <= r; i++ {
			if cmp(v.Index(i).Interface(), v.Index(l).Interface()) {
				j++
				swap(v.Index(i), v.Index(j))
			}
		}
		swap(v.Index(l), v.Index(j))
		qsort(l, j-1)
		qsort(j+1, r)
	}
	qsort(0, v.Len()-1)
}
