package main

import (
	"bytes"
	"compress/gzip"
	"github.com/nokivan/gravatar"
	"io"
	"net/http"
	"strings"
	"time"
)

var indexCache *bytes.Buffer
var feedCache *bytes.Buffer

func updateIndexAndFeed() {
	// TODO pager
	indexList := getArticleList()
	sortSlice(indexList, func(a, b interface{}) bool {
		return a.(*Article).Id > b.(*Article).Id
	})
	for i, _ := range indexList {
		temp := *indexList[i]
		temp.Content = makeSummary(temp.Content)
		indexList[i] = &temp
	}
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
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Write(indexCache.Bytes())
}

func feedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Write(feedCache.Bytes())
}

func initIndex() {
	if config["RootUrl"] != "/" {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, config["RootUrl"], http.StatusFound)
		})
	}
	http.HandleFunc(config["RootUrl"], getGzipHandler(indexHandler))
	http.HandleFunc(config["RootUrl"]+"feed", getGzipHandler(feedHandler))
}

type responseRewriter struct {
	http.ResponseWriter
	io.Writer
}

func (this responseRewriter) Write(data []byte) (int, error) {
	return this.Writer.Write(data)
}

func handler2HandlerFunc(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

func getGzipHandler(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			f(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gw := gzip.NewWriter(w)
		defer gw.Close()
		f(responseRewriter{
			ResponseWriter: w,
			Writer:         gw,
		}, r)
	}
}

func makeSummary(s string) string {
	min := 0
	for min < 600 && s[min:] != "" {
		if idx := strings.Index(s[min:], "</p>"); idx != -1 {
			min += idx + len("</p>")
			continue
		}
		if idx := strings.Index(s[min:], "<br/>"); idx != -1 {
			min += idx + len("<br/>")
			continue
		}
		if idx := strings.Index(s[min:], "<br>"); idx != -1 {
			min += idx + len("<br>")
			continue
		}
		min = len(s)
	}
	if min != len(s) {
		return s[:min] + "<br/>... ...<br/>"
	}
	return s[:min]
}

func getGravatarURL(email string, size int) string {
	emailHash := gravatar.EmailHash(email)
	url := gravatar.GetAvatarURL("https", emailHash, gravatar.DefaultMonster, size)
	return url.String()
}
