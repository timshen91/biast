package main

import (
	"errors"
	"exp/html"
	"exp/html/atom"
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
	atom.Q:          map[string]struct{}{"cite": struct{}{}},
	atom.Strike:     map[string]struct{}{},
	atom.Strong:     map[string]struct{}{},
	// other tags?
}

func articleHandler(w http.ResponseWriter, r *http.Request) {
	id64, err := strconv.ParseUint(r.URL.Path[len(config["ArticleUrl"]):], 10, 32)
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
		comm, err := genComment(r, id)
		if err != nil {
			logger.Println(r.RemoteAddr+":", err.Error())
			feedback = "Oops...! " + err.Error()
		}
		// EventStart: newComment
		artMgr.atomAppendComment(comm)
		db.sync(commentPrefix, comm)
		go newCommentNotify(comm)
		// EventEnd: newComment
	}
	tmpl.ExecuteTemplate(w, "article", map[string]interface{}{
		"config":   config,
		"article":  p,
		"comments": artMgr.atomGetCommentList(p.Id),
		"feedback": feedback,
	})
}

func genComment(r *http.Request, fid aid) (*Comment, error) {
	if !checkKeyExist(r.Form, "author", "email", "content", "notify") {
		return nil, errors.New("required field not found")
	}
	if len(r.Form.Get("author")) == 0 || len(r.Form.Get("email")) == 0 {
		return nil, errors.New("name, email and content can't be blank")
	}
	content, err := tagFilter(r.Form.Get("content"))
	if err != nil {
		return nil, err
	}
	return &Comment{
		Id:         artMgr.allocCommentId(),
		Author:     html.EscapeString(r.Form.Get("author")),
		Email:      html.EscapeString(r.Form.Get("email")),
		RemoteAddr: r.RemoteAddr,
		Date:       time.Now(),
		Content:    content,
		Father:     fid,
		ReplyNotif: r.Form.Get("notify") == "on",
	}, nil
}

func init() {
	config["ArticleUrl"] = config["RootUrl"] + "article/"
	http.HandleFunc(config["ArticleUrl"], articleHandler)
}

func tagFilter(content string) (string, error) {
	var ret string
	t := html.NewTokenizer(strings.NewReader(content))
	stack := make([]atom.Atom, 0)
L:
	for {
		t.Next()
		token := t.Token()
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
				ret += token.String()
			} else {
				ret += html.EscapeString(token.String())
			}
		case html.EndTagToken:
			var top int = len(stack) - 1
			for top >= 0 && stack[top] != token.DataAtom {
				top--
			}
			if top == -1 {
				ret += html.EscapeString(token.String())
			} else {
				stack = stack[0:top]
				ret += token.String()
			}
		case html.TextToken:
			ret += html.EscapeString(token.String())
		case html.ErrorToken:
			break L
		}
	}
	if err := t.Err(); err != io.EOF {
		return "", err
	}
	return ret, nil
}
