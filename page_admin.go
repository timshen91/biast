package main

import (
	"errors"
	"html"
	"net/http"
	"time"
)

func newArticle(w http.ResponseWriter, r *http.Request) {
	var feedback string
	article := &Article{}
	if r.Method == "POST" {
		feedback = "Article sent"
		r.ParseForm()
		var err error
		if article, err = genArticle(r); err != nil {
			feedback = "Oops...! " + err.Error()
		}
		// EventStart: newArticle
		artMgr.atomSetArticle(article)
		db.sync(articlePrefix, article)
		go updateIndex()
		// EventEnd: newArticle
	}
	tmpl.ExecuteTemplate(w, "new", map[string]interface{}{
		"config":   config,
		"feedback": feedback,
		"form":     article,
	})
}

func genArticle(r *http.Request) (*Article, error) {
	if !checkKeyExist(r.Form, "author", "email", "content", "title") {
		logger.Println("new:", "required field not exists")
		return nil, errors.New("required field not exists")
	}
	// may we need a filter?
	return &Article{
		Id:         artMgr.allocArticleId(),
		Author:     html.EscapeString(r.Form.Get("author")),
		Email:      html.EscapeString(r.Form.Get("email")),
		RemoteAddr: r.RemoteAddr,
		Date:       time.Now(),
		Title:      html.EscapeString(r.Form.Get("title")),
		Content:    r.Form.Get("content"),
	}, nil
}

func init() {
	if config["AdminUrl"][len(config["AdminUrl"])-1] == '/' {
		config["AdminUrl"] = config["AdminUrl"][:len(config["AdminUrl"])-1]
	}
	http.HandleFunc(config["RootUrl"]+config["AdminUrl"], newArticle)
	// http.HandleFunc(config["AdminUrl"] + "modify", modifyArticle)
}
