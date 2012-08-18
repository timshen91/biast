package main

import (
	"net/http"
	"sort"
)

type listForSort []*article

func (this listForSort) Len() int {
	return len(this)
}

func (this listForSort) Less(i, j int) bool {
	return this[i].date > this[j].date // from lastest to the oldest
}

func (this listForSort) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

var indexList listForSort

func updateIndex() {
	indexList = make([]*article, 16)
	for _, p := range articles {
		if p != nil {
			indexList = append(indexList, p)
		}
	}
	sort.Sort(indexList)
}

func indexHandler(w http.ResponseWriter, r * http.Request) {
	// TODO we just need part of articles instead of the whole, say, from 10 to 20
}

func indexInit() {
	http.HandleFunc(config["rootUrl"], indexHandler)
	updateIndex()
}
