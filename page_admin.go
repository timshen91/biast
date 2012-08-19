package main

import (
	"errors"
	"net/http"
	"time"
)

func allocId() int {
	return 0
}

func newArticle(w http.ResponseWriter, r *http.Request) {
	var feedback string
	var form *Article
	if r.Method == "POST" {
		feedback = "Article sent"
		r.ParseForm()
		if err := func() error {
			if !checkKeyExist(r.Form, "author", "email", "content", "title") {
				logger.Println("new:", "required field not exists")
				return errors.New("required field not exists")
			}
			id := allocId()
			form = &Article{
				info: info{
					Author:     r.Form.Get("author"),
					Email:      r.Form.Get("email"),
					Content:    r.Form.Get("content"),
					RemoteAddr: r.RemoteAddr,
					Date:       time.Now(),
				},
				Id:       id,
				Title:    r.Form.Get("title"),
				Comments: make([]*Comment, 0),
			}
			articles[id] = form // FIXME it may not be atomic
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
	http.HandleFunc(config["AdminUrl"]+"new", newArticle)
	// http.HandleFunc(config["AdminUrl"] + "modify", modifyArticle)
}
