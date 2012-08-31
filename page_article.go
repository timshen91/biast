package main

import (
	"errors"
	"exp/html"
	"exp/html/atom"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var validAtom = map[atom.Atom]map[string]struct{}{
	atom.A:          map[string]struct{}{"href": struct{}{}, "title": struct{}{}},
	atom.Abbr:       map[string]struct{}{"title": struct{}{}},
	atom.B:          map[string]struct{}{},
	atom.Blockquote: map[string]struct{}{"cite": struct{}{}},
	atom.Body:       map[string]struct{}{},
	atom.Cite:       map[string]struct{}{},
	atom.Code:       map[string]struct{}{},
	atom.Del:        map[string]struct{}{"detetime": struct{}{}},
	atom.Em:         map[string]struct{}{},
	atom.I:          map[string]struct{}{},
	atom.P:          map[string]struct{}{},
	atom.Pre:        map[string]struct{}{},
	atom.Q:          map[string]struct{}{"cite": struct{}{}},
	atom.Strike:     map[string]struct{}{},
	atom.Strong:     map[string]struct{}{},
	// other tags?
}

func articleHandler(w http.ResponseWriter, r *http.Request) {
	id64, err := strconv.ParseUint(r.URL.Path[len(config["ArticleUrl"]):], 10, 32)
	if err != nil {
		logger.Println(r.RemoteAddr + ": 404 for an invalid id")
		http.NotFound(w, r)
		return
	}
	id := aid(id64)
	p := getArticle(id)
	if p == nil {
		logger.Println(r.RemoteAddr + "404")
		http.NotFound(w, r)
		return
	}
	var feedback string
	if r.Method == "POST" {
		comm, err := genComment(r, id)
		if err != nil {
			logger.Println(r.RemoteAddr+":", err.Error())
			feedback = "Oops...! " + err.Error()
		} else {
			// EventStart: newComment
			appendComment(comm)
			db.sync(commentPrefix, fmt.Sprint(comm.Id), comm)
			go newCommentNotify(comm)
			// EventEnd: newComment
		}
	}
	if err := tmpl.ExecuteTemplate(w, "article", map[string]interface{}{
		"config":   config,
		"article":  p,
		"comments": getCommentList(p.Id),
		"feedback": feedback,
		"header":   p.Title,
	}); err != nil {
		logger.Println(r.RemoteAddr + err.Error())
	}
}

func genComment(r *http.Request, fid aid) (*Comment, error) {
	r.ParseForm()
	if !checkKeyExist(r.Form, "author", "email", "content") {
		return nil, errors.New("required field not found")
	}
	if len(r.Form.Get("author")) == 0 || len(r.Form.Get("email")) == 0 {
		return nil, errors.New("name, email and content can't be blank")
	}
	content, err := htmlFilter(r.Form.Get("content"))
	if err != nil {
		return nil, err
	}
	if len(r.Form.Get("content")) == 0 {
		return nil, errors.New("name, email and content can't be blank")
	}
	return &Comment{
		Id:         allocCommentId(),
		Author:     html.EscapeString(r.Form.Get("author")),
		Email:      html.EscapeString(r.Form.Get("email")),
		Website:    genWebsite(r.Form.Get("website")),
		RemoteAddr: r.RemoteAddr,
		Date:       time.Now(),
		Content:    content,
		Father:     fid,
		QuoteNotif: r.Form.Get("notify") == "on",
	}, nil
}

func init() {
	config["ArticleUrl"] = config["RootUrl"] + "article/"
	http.HandleFunc(config["ArticleUrl"], articleHandler)
}

func htmlFilter(content string) (string, error) {
	var ret string
	t := html.NewTokenizer(strings.NewReader(content))
	stack := make([]atom.Atom, 0)
L:
	for {
		t.Next()
		token := t.Token()
		str := html.UnescapeString(token.String())
		switch token.Type {
		case html.StartTagToken, html.SelfClosingTagToken:
			ans := false
			if attrMap, ex := validAtom[token.DataAtom]; ex {
				ans = true
				for _, attr := range token.Attr {
					if _, ex := attrMap[attr.Key]; !ex {
						ans = false
						break
					}
				}
			}
			if ans {
				stack = append(stack, token.DataAtom)
				ret += str
			} else {
				ret += html.EscapeString(str)
			}
		case html.EndTagToken:
			var top int = len(stack) - 1
			for top >= 0 && stack[top] != token.DataAtom {
				top--
			}
			if top == -1 {
				ret += html.EscapeString(str)
			} else {
				stack = stack[0:top]
				ret += str
			}
		case html.TextToken:
			ret += html.EscapeString(str)
		case html.ErrorToken:
			break L
		}
	}
	if err := t.Err(); err != io.EOF {
		return "", err
	}
	return ret, nil
}

func genWebsite(url string) string {
	if url != "" && !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "http://" + url
	}
	return url
}
