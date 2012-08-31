package main

import (
	"bytes"
	"net/http"
	"time"
)

var indexCache *bytes.Buffer
var feedCache *bytes.Buffer

func updateIndexAndFeed() {
	// TODO pager
	indexList := getArticleList()
	qsortForArticleList(indexList, 0, len(indexList)-1)
	// index
	newIndexCache := &bytes.Buffer{}
	if err := tmpl.ExecuteTemplate(newIndexCache, "index", map[string]interface{}{
		"config":   config,
		"articles": indexList,
		"header":   config["ServerName"],
	}); err != nil {
		logger.Println("index cache:", err.Error())
	}
	indexCache = newIndexCache
	newFeedCache := &bytes.Buffer{}
	if err := tmpl.ExecuteTemplate(newFeedCache, "feed", map[string]interface{}{
		"config":    config,
		"articles":  indexList,
		"lastBuild": time.Now().String(),
	}); err != nil {
		logger.Println("feed cache:", err.Error())
	}
	feedCache = newFeedCache
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(indexCache.Bytes())
}

func feedHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(feedCache.Bytes())
}

func init() {
	http.HandleFunc(config["RootUrl"], indexHandler)
	http.HandleFunc(config["RootUrl"]+"feed", feedHandler)
}

func qsortForArticleList(a []*Article, l, r int) {
	if l > r {
		return
	}
	i := l
	j := (r-l)/2 + l
	a[i], a[j] = a[j], a[i]
	j = l
	for i = l + 1; i <= r; i++ {
		if a[i].Id > a[l].Id {
			j++
			a[j], a[i] = a[i], a[j]
		}
	}
	a[j], a[l] = a[l], a[j]
	qsortForArticleList(a, l, j-1)
	qsortForArticleList(a, j+1, r)
}
