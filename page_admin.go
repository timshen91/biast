package main

import (
	"errors"
	"html"
	"net/http"
	"time"
)

func newArticle(w http.ResponseWriter, r *http.Request) {
	var feedback string
	form := &Article{}
	if r.Method == "POST" {
		feedback = "Article sent"
		r.ParseForm()
		if err := func() error {
			if !checkKeyExist(r.Form, "author", "email", "content", "title") {
				logger.Println("new:", "required field not exists")
				return errors.New("required field not exists")
			}
			form = &Article{
				Author:     html.EscapeString(r.Form.Get("author")),
				Email:      html.EscapeString(r.Form.Get("email")),
				RemoteAddr: r.RemoteAddr,
				Date:       time.Now(),
				Title:      html.EscapeString(r.Form.Get("title")),
				Content:    r.Form.Get("content"),
			}
			id := artMgr.allocArticleId()
			form.Id = id
			// EventStart: newArticle
			artMgr.atomSetArticle(form)
			db.sync(articlePrefix, form)
			go updateIndex()
			// EventEnd: newArticle
			return nil
		}(); err != nil {
			feedback = "Oops...! " + err.Error()
		}
	}
	tmpl.ExecuteTemplate(w, "new", map[string]interface{}{
		"config":   config,
		"feedback": feedback,
		"form":     form,
	})
}

func initPageAdmin() {
	if config["AdminUrl"][len(config["AdminUrl"])-1] != '/' {
		config["AdminUrl"] += "/"
	}
	http.HandleFunc(config["RootUrl"]+config["AdminUrl"]+"new", newArticle)
	// http.HandleFunc(config["AdminUrl"] + "modify", modifyArticle)
}
