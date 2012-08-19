package main

import (
	"net/http"
	"strconv"
	"errors"
	"time"
)

func articleHandler(w http.ResponseWriter, r * http.Request) {
	idStr := r.URL.Path[len(config["articleUrl"]):] // FIXME don't use fixed length
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
			newComm := &comment{
				info : info{
					Author : r.Form.Get("author"),
					Email : r.Form.Get("email"),
					Content : r.Form.Get("content"),
					RemoteAddr : r.RemoteAddr,
					Date : time.Now(),
				},
				Father : id,
			}
			if ok := func(comm *comment) bool {
				return true
			}(newComm); !ok {
				// TODO comment filter
			}
			commList := p.Comments
			commList = append(commList, newComm)
			return nil
		}(); err != nil {
			logger.Println(r.RemoteAddr + ":", err.Error())
			feedback = err.Error()
		}
	}
	tmpl.ExecuteTemplate(w, "article", map[string]interface{}{
		"config": config,
		"article": p,
		"feedback": feedback,
	})
}

func articleInit() {
    config["articleUrl"] = config["rootUrl"] + "article/"
    http.HandleFunc(config["articleUrl"], articleHandler)
}
