package main

import (
	"crypto/md5"
	"errors"
	"exp/html"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

func newArticleHandler(w http.ResponseWriter, r *http.Request) {
	var feedback string
	var article = &Article{}
	r.ParseForm()
	old, ok := getOld(r.URL.Path)
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
				feedback = "Oops..! " + "This email address can't be the author."
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
			article = old
		}
	}
out:
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	tagsNow := strings.Join(article.Tags, ", ")
	if err := tmpl.ExecuteTemplate(w, "new", map[string]interface{}{
		"config":   config,
		"feedback": feedback,
		"form": map[string]string{
			"Title":   html.EscapeString(article.Title),
			"Author":  html.EscapeString(article.Author),
			"Website": html.EscapeString(article.Website),
			"Content": html.EscapeString(article.Src),
		},
		"tagsNow": tagsNow,
		"allTags": getAllTags(),
		"header":  "Admin Panel",
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

func newArticleAuth(a, old *Article) {
	url := notifRegister(func(w http.ResponseWriter, r *http.Request) {
		if old != nil {
			go updateTags(a.Id, old.Tags, a.Tags)
		} else {
			go updateTags(a.Id, nil, a.Tags)
		}
		setArticle(a)
		go updateIndexAndFeed()
		http.Redirect(w, r, config["ArticleUrl"]+fmt.Sprint(a.Id), http.StatusFound)
	})
	send(a.Email, "New article authentication", fmt.Sprintf(`<p>Dear %s, you have an article to be authenticated for publishment:
<p>
%s
</p></p>
<p>If you know what you are doing, please click <a href="%s">here</a>.</p>`, a.Author, a.Content, url))
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
	sortSlice(ret, func(a, b interface{}) bool {
		return a.(string) < b.(string)
	})
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
		Src:        r.Form.Get("content"),
		Content:    process(r.Form.Get("content")),
		Notif:      r.Form.Get("notify") == "on",
		Tags:       tagList,
	}, nil
}

func process(content string) string {
	ret := ""
	t := html.NewTokenizer(strings.NewReader(content))
	latex := false
	latexSrc := ""
L:
	for {
		t.Next()
		token := t.Token()
		str := html.UnescapeString(token.String())
		switch token.Type {
		case html.StartTagToken:
			if token.Data == "latex" {
				latex = true
			}
		case html.ErrorToken:
			break L
		case html.EndTagToken:
			if token.Data == "latex" {
				latex = false
				ret += fmt.Sprintf("<img src=\"%s\" alt=\"%s\"/>", genLaTeX(latexSrc), html.EscapeString(latexSrc))
			}
		default:
			if latex {
				latexSrc += str
			} else {
				ret += str
			}
		}
	}
	return ret
}

var latexMutex sync.Mutex

func genLaTeX(src string) string {
	println(src)
	cmd := exec.Command("/usr/bin/latex", "-output-directory=/tmp")
	cmd.Stdin = strings.NewReader(fmt.Sprintf(`
\documentclass{article}
\pagestyle{empty}

\begin{document}
{\huge %s}
\end{document}`, src))
	latexMutex.Lock()
	defer latexMutex.Unlock()
	if err := cmd.Run(); err != nil {
		logger.Println("latex:", err.Error())
		return ""
	}
	fileName := getLaTeXFileName(src)
	println(config["DocumentRoot"] + "static/image/" + fileName)
	if err := exec.Command("/usr/bin/convert", "-trim", "/tmp/texput.dvi", config["DocumentPath"]+"static/image/"+fileName).Run(); err != nil {
		logger.Println("latex:", err.Error())
		return ""
	}
	return config["RootUrl"] + "image/" + fileName
}

func getLaTeXFileName(src string) string {
	h := md5.New()
	io.WriteString(h, src)
	return fmt.Sprintf("%x", h.Sum(nil)) + ".png" // FIXME collision
}

func checkArticle(a *Article) error {
	if a.Author == "" || a.Email == "" || a.Content == "" || a.Title == "" {
		return errors.New("Name, email, content and title can't be blank.")
	}
	if _, ex := adminList[a.Email]; !ex {
		return errors.New("This email address can't be the author.")
	}
	return nil
}

func initPageAdmin() {
	config["AdminUrl"] = config["RootUrl"] + "admin/"
	http.HandleFunc(config["AdminUrl"], getGzipHandler(newArticleHandler))
}
