package main

import (
	"errors"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newArticleHandler(w http.ResponseWriter, r *http.Request) {
	var feedback string
	var article = &Article{}
	idRequest, ok := parseId(r.URL.Path)
	if r.Method == "POST" {
		if temp, err := genArticle(r); err != nil {
			feedback = "Oops...! " + err.Error()
			goto out
		} else {
			article = temp
		}
		if err := checkArticle(article); err != nil {
			feedback = "Oops...! " + err.Error()
			goto out
		}
		old := getArticle(article.Id)
		if old != nil && old.Email != article.Email {
			feedback = "Oops..! " + "Only the author can modify its article."
			goto out
		}
		if old != nil {
			article.Date = old.Date
		} else {
			article.Date = time.Now()
		}
		if ok {
			article.Id = idRequest
		} else {
			article.Id = allocArticleId()
		}
		// EventStart: newArticle
		newArticleAuth(article, old)
		// EventEnd: newArticle
		feedback = "Welcome my dear admin! Mail request has been sent, please check your mail."
	} else {
		if ok {
			if temp := getArticle(idRequest); temp != nil {
				article = temp
			}
		}
	}
out:
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	tagsNow := strings.Join(article.Tags, ", ")
	if err := tmpl.ExecuteTemplate(w, "new", map[string]interface{}{
		"config":   config,
		"feedback": feedback,
		"form":     article,
		"tagsNow":  tagsNow,
		"allTags":  getAllTags(),
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

func genTags(tagList string) []string { // tags shouldn't contain quote marks, please don't ask why...
	m := map[string]struct{}{}
	for _, s := range strings.Split(tagList, ",") {
		strings.Replace(s, "'", "", -1)
		strings.Replace(s, "\"", "", -1)
		if t := strings.TrimSpace(s); len(t) != 0 {
			m[t] = struct{}{}
		}
	}
	ret := []string{}
	for t, _ := range m {
		ret = append(ret, t)
	}
	return ret
}

func genArticle(r *http.Request) (*Article, error) {
	r.ParseForm()
	if !checkKeyExist(r.Form, "author", "email", "content", "title") {
		logger.Println("new:", "required field not exists.")
		return nil, errors.New("Required field not exists.")
	}
	tagList := genTags(r.Form.Get("tags"))
	// may we need a filter?
	return &Article{
		Author:     html.EscapeString(r.Form.Get("author")),
		Email:      html.EscapeString(r.Form.Get("email")),
		Website:    genWebsite(r.Form.Get("website")),
		RemoteAddr: r.RemoteAddr,
		Date:       time.Now(),
		Title:      html.EscapeString(r.Form.Get("title")),
		Content:    r.Form.Get("content"),
		Notif:      r.Form.Get("notify") == "on",
		Tags:       tagList,
	}, nil
}

func checkArticle(a *Article) error {
	if a.Author == "" || a.Email == "" || a.Content == "" || a.Title == "" {
		return errors.New("Name, email, content and title can't be blank.")
	}
	if _, ex := adminList[a.Email]; !ex {
		return errors.New("This email is not registered as an admin.")
	}
	return nil
}

func initPageAdmin() {
	config["AdminUrl"] = config["RootUrl"] + "admin/"
	http.HandleFunc(config["AdminUrl"], getGzipHandler(newArticleHandler))
}
