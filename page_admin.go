package main

import (
	"errors"
	"html"
	"net/http"
	"strconv"
	"time"
)

func newArticle(w http.ResponseWriter, r *http.Request) {
	var feedback string
	var article *Article
	idRequest, ok := parseId(r.URL.Path)
	if r.Method == "POST" {
		feedback = "Article sent"
		var err error
		if article, err = genArticle(r); err != nil {
			article = &Article{}
			feedback = "Oops...! " + err.Error()
		} else {
			if ok {
				article.Id = idRequest
			} else {
				article.Id = artMgr.allocArticleId()
			}
			// EventStart: newArticle
			artMgr.setArticle(article)
			db.sync(articlePrefix, article)
			go updateIndexAndFeed()
			// EventEnd: newArticle
		}
	} else {
		article = &Article{}
		if ok {
			if temp := artMgr.getArticle(idRequest); temp != nil {
				article = temp
			}
		}
	}

	if err := tmpl.ExecuteTemplate(w, "new", map[string]interface{}{
		"config":   config,
		"feedback": feedback,
		"form":     article,
		"header":   "Admin Panel",
	}); err != nil {
		logger.Println("new:", err.Error())
	}
}

func parseId(url string) (id aid, ok bool) {
	if url[len(config["AdminUrl"]):] == "" {
		return 0, false
	}
	id64, err := strconv.ParseUint(url[len(config["AdminUrl"]):], 10, 32)
	if err != nil {
		return 0, false
	}
	return aid(id64), true
}

func genArticle(r *http.Request) (*Article, error) {
	r.ParseForm()
	if !checkKeyExist(r.Form, "author", "email", "content", "title") {
		logger.Println("new:", "required field not exists")
		return nil, errors.New("required field not exists")
	}
	if r.Form.Get("author") == "" || r.Form.Get("email") == "" || r.Form.Get("content") == "" || r.Form.Get("title") == "" {
		return nil, errors.New("name, email, content and title can't be blank")
	}
	// may we need a filter?
	return &Article{
		Author:     html.EscapeString(r.Form.Get("author")),
		Email:      html.EscapeString(r.Form.Get("email")),
		Website:    genWebsite(r.Form.Get("website")),
		RemoteAddr: r.RemoteAddr,
		Date:       time.Now(),
		Title:      html.EscapeString(r.Form.Get("title")),
		Content:    r.Form.Get("content"),
		QuoteNotif: r.Form.Get("notify") == "on",
	}, nil
}

func init() {
	if config["AdminUrl"][len(config["AdminUrl"])-1] != '/' {
		config["AdminUrl"] += "/"
	}
	if config["AdminUrl"][0] == '/' {
		config["AdminUrl"] = config["AdminUrl"][1:]
	}
	config["AdminUrl"] = config["RootUrl"] + config["AdminUrl"]
	http.HandleFunc(config["AdminUrl"], newArticle)
}
