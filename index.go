package main

import (
	"net/http"
	"sort"
	"bytes"
)

type listForSort []*article

func (this listForSort) Len() int {
	return len(this)
}

func (this listForSort) Less(i, j int) bool {
	return this[i].date.Unix() > this[j].date.Unix() // from lastest to the oldest
}

func (this listForSort) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

var indexList listForSort
var indexCache bytes.Buffer

func updateIndex() {
	// TODO pager
	indexList = make([]*article, 16)
	for _, p := range articles {
		if p != nil {
			indexList = append(indexList, p)
		}
	}
	sort.Sort(indexList)
	indexCache.Reset()
	tmpl.ExecuteTemplate(&indexCache, "index", map[string]interface{}{"config": config, "articles": indexList})
}

func indexHandler(w http.ResponseWriter, r * http.Request) {
	indexCache.WriteTo(w)
}

func indexInit() {
	http.HandleFunc(config["rootUrl"], indexHandler)
	updateIndex()
}
