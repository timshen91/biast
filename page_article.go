package main

import (
	"errors"
	"html"
	"net/http"
	"strconv"
	"time"
)

func articleHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len(config["ArticleUrl"]):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Println(r.RemoteAddr + ": 404 for an invalid id")
		http.NotFound(w, r)
		return
	}
	p := articles[id]
	if p == nil {
		logger.Println("404 for an nonexist id")
		http.NotFound(w, r)
		return
	}
	var feedback string
	if r.Method == "POST" {
		r.ParseForm()
		if err := func() error {
			if !checkKeyExist(r.Form, "author", "email", "content") {
				return errors.New("required field not found")
			}
			newComm := &Comment{
				info: info{
					Author:     html.EscapeString(r.Form.Get("author")),
					Email:      html.EscapeString(r.Form.Get("email")),
					Content:    html.EscapeString(r.Form.Get("content")),
					RemoteAddr: r.RemoteAddr,
					Date:       time.Now(),
				},
				Father: id,
			}
			if ok := func(comm *Comment) bool {
				return true
			}(newComm); !ok {
				// TODO comment filter
			}
			commList := p.Comments
			commList = append(commList, newComm)
			return nil
		}(); err != nil {
			logger.Println(r.RemoteAddr+":", err.Error())
			feedback = err.Error()
		}
	}
	tmpl.ExecuteTemplate(w, "article", map[string]interface{}{
		"config":   config,
		"article":  p,
		"feedback": feedback,
	})
}

func initPageArticle() {
	config["ArticleUrl"] = config["RootUrl"] + "article/"
	http.HandleFunc(config["ArticleUrl"], articleHandler)
}
