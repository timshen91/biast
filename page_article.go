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
	id64, err := strconv.ParseUint(idStr, 10, 32)
	id := aid(id64)
	if err != nil {
		logger.Println(r.RemoteAddr + ": 404 for an invalid id")
		http.NotFound(w, r)
		return
	}
	p := artMgr.atomGetArticle(id)
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
				Author:     html.EscapeString(r.Form.Get("author")),
				Email:      html.EscapeString(r.Form.Get("email")),
				RemoteAddr: r.RemoteAddr,
				Date:       time.Now(),
				Father:     id,
				Content:    html.EscapeString(r.Form.Get("content")),
				ReplyNotif: r.Form.Get("notify") == "on",
			}
			if ok := func(comm *Comment) bool {
				if len(comm.Author) != 0 &&
					len(comm.Email) != 0 &&
					len(comm.Content) != 0 {
					return true
				}
				feedback = "name, email and content can't be blank"
				return false
			}(newComm); !ok {
				return nil
			}
			newComm.Id = artMgr.allocCommentId()
			// EventStart: newComment
			artMgr.atomAppendComment(newComm)
			db.sync(commentPrefix, newComm)
			go newCommentNotify(newComm)
			// EventEnd: newComment
			return nil
		}(); err != nil {
			logger.Println(r.RemoteAddr+":", err.Error())
			feedback = err.Error()
		}
	}
	tmpl.ExecuteTemplate(w, "article", map[string]interface{}{
		"config":   config,
		"article":  p,
		"comments": artMgr.atomGetCommentList(p.Id),
		"feedback": feedback,
	})
}

func initPageArticle() {
	config["ArticleUrl"] = config["RootUrl"] + "article/"
	http.HandleFunc(config["ArticleUrl"], articleHandler)
}
