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
	old, ok := getOld(r.URL.Path)
	r.ParseForm()
	if r.Method == "POST" {
		if temp, err := genArticle(r); err != nil {
			feedback = "Oops...! " + err.Error()
			goto out
		} else {
			article = temp
		}
		if r.Form.Get("post") == "preview" {
			w.Header().Set("Content-Type", "text/html; charset=UTF-8")
			if err := tmpl.ExecuteTemplate(w, "preview", map[string]interface{}{
				"config":  config,
				"article": article,
				"header":  article.Title + "(Preview)",
			}); err != nil {
				logger.Println("new:", err.Error())
			}
			return
		}
		if err := checkArticle(article); err != nil {
			feedback = "Oops...! " + err.Error()
			goto out
		}
		if ok {
			if old.Email != article.Email {
				feedback = "Oops..! " + "Only the author can modify its article."
				goto out
			}
			article.Id = old.Id
			article.Date = old.Date
		} else {
			article.Id = allocArticleId()
		}
		// EventStart: newArticle
		newArticleAuth(article, old)
		// EventEnd: newArticle
		feedback = "Welcome my dear admin! Mail request has been sent, please check your mail."
	} else {
		if ok {
			old.Email = ""
			article = old
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

func getOld(url string) (old *Article, ex bool) {
	if url[len(config["AdminUrl"]):] == "" {
		return nil, false
	}
	id64, err := strconv.ParseUint(url[len(config["AdminUrl"]):], 10, 32)
	if err != nil {
		return nil, false
	}
	ret := getArticle(aid(id64))
	if ret == nil {
		return nil, false
	}
	return ret, true
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
