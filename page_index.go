package main

import (
	"bytes"
	"net/http"
	"sort"
)

type listForSort []*Article

func (this listForSort) Len() int {
	return len(this)
}

func (this listForSort) Less(i, j int) bool {
	return this[i].Date.Unix() > this[j].Date.Unix() // from lastest to the oldest
}

func (this listForSort) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

var indexList listForSort
var indexCache bytes.Buffer

func updateIndex() {
	// TODO pager
	indexList = make([]*Article, 16)
	for _, p := range articles {
		if p != nil {
			indexList = append(indexList, p)
		}
	}
	sort.Sort(indexList)
	indexCache.Reset()
	tmpl.ExecuteTemplate(&indexCache, "index", map[string]interface{}{
		"config":   config,
		"articles": indexList,
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	indexCache.WriteTo(w)
}

func initPageIndex() {
	http.HandleFunc(config["RootUrl"], indexHandler)
	updateIndex()
}
