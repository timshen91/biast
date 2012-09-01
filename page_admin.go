package main

import (
	"errors"
	"fmt"
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
		temp, err := genArticle(r);
		if err == nil {
			article = temp
			if ok {
				article.Id = idRequest
			} else {
				article.Id = allocArticleId()
				article.Date = time.Now()
			}
			old := getArticle(article.Id)
			if old != nil {
				article.Date = old.Date
			} else {
				article.Date = time.Now()
			}
			// EventStart: newArticle
			if old != nil {
				go updateTags(article.Id, old.Tags, article.Tags)
			} else {
				go updateTags(article.Id, nil, article.Tags)
			}
			setArticle(article)
			go updateIndexAndFeed()
			// EventEnd: newArticle
			http.Redirect(w, r, config["ArticleUrl"]+fmt.Sprint(article.Id), http.StatusFound)
			return
		}
		feedback = "Oops...! " + err.Error()
	} else {
		if ok {
			if temp := getArticle(idRequest); temp != nil {
				article = temp
			}
		}
	}
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

func genTags(tagList string) []string { // tags shouldn't contain quote marks
	ret := []string{}
	for _, s := range strings.Split(tagList, ",") {
		strings.Replace(s, "'", "", -1)
		strings.Replace(s, "\"", "", -1)
		if t := strings.TrimSpace(s); len(t) != 0 {
			ret = append(ret, t)
		}
	}
	return ret
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
	var tagList = []string{}
	for _, t := range genTags(r.Form.Get("tags")) {
		tagList = append(tagList, t)
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
		Notif:      r.Form.Get("notify") == "on",
		Tags:       tagList,
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
	http.HandleFunc(config["AdminUrl"], newArticleHandler)
}
