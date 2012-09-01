package main

import (
	"exp/html"
	"exp/html/atom"
	"fmt"
	"math/rand"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

func newCommentNotify(comm *Comment) {
	// notify the article author
	if father := getArticle(comm.Father); father.Notif {
		send(father.Email, "Your article has been commented", fmt.Sprintf(
			`Dear %s, your article on %s has been commented by %s:
	<blockquote>%s</blockquote>
	Click <a href="%s">here</a> for details. Click <a href="%s">here</a> to close the notification.
	`, father.Author, config["ServerName"], comm.Author, comm.Content, config["Domain"]+config["ArticleUrl"]+fmt.Sprint(father.Id)+"#comment-"+fmt.Sprint(comm.Id), "http://"+config["Domain"]+config["ResponseUrl"]+allocRandomKey(getCloseArticleNotif(comm.Father))))
	}
	// notify the commenter
	fmt.Println(parseRef(comm.Content))
	for _, id := range parseRef(comm.Content) {
		if p := getComment(id); p != nil && p.Notif {
			if comm.Father == p.Father {
				send(p.Email, "Your comment has been quoted", fmt.Sprintf(
					`Dear %s, your comment on %s has been quoted by %s:
<blockquote>%s</blockquote>
Click <a href="%s">here</a> for details. Click <a href="%s">here</a> to close the notification.
`, p.Author, config["ServerName"], comm.Author, comm.Content, config["Domain"]+config["ArticleUrl"]+fmt.Sprint(p.Father)+"#comment-"+fmt.Sprint(comm.Id), "http://"+config["Domain"]+config["ResponseUrl"]+allocRandomKey(getCloseCommentNotif(id))))
			}
		}
	}
}

func getCloseArticleNotif(id aid) func() {
	return func() {
		article := getArticle(id)
		article.Notif = false
		setArticle(article)
	}
}

func getCloseCommentNotif(id cid) func() {
	return func() {
		comment := getComment(id)
		comment.Notif = false
		setComment(comment)
	}
}

func allocRandomKey(callback func()) string {
	var key string
	for {
		key = fmt.Sprint(rand.Uint32())
		if _, ex := responseCallback[key]; !ex {
			responseCallback[key] = callback
			break
		}
	}
	return fmt.Sprint(key)
}

func mailResponseHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len(config["ResponseUrl"]):]
	callback, ex := responseCallback[key]
	if ex {
		callback()
		delete(responseCallback, key)
		w.Write([]byte(`Success`))
	} else {
		w.Write([]byte(`Invalid key`))
	}
}

var mailAuth smtp.Auth
var responseCallback = map[string]func(){}

func init() {
	if _, ok := config["SMTPUsername"]; !ok {
		config["SMTPUsername"] = ""
	}
	if _, ok := config["SMTPPass"]; !ok {
		config["SMTPPass"] = ""
	}
	if _, ok := config["SMTPAddr"]; !ok {
		config["SMTPAddr"] = "127.0.0.1:25"
	}
	mailAuth = smtp.PlainAuth("", config["SMTPUsername"], config["SMTPPass"], config["SMTPAddr"])
	rand.Seed(time.Now().Unix())
	config["ResponseUrl"] = config["RootUrl"] + "response/"
	http.HandleFunc(config["ResponseUrl"], getGzipHandler(mailResponseHandler))
}

func send(to, subject, msg string) {
	logger.Println("notification:", "SendMail:", msg, "for", to)
	if err := smtp.SendMail(config["SMTPAddr"], mailAuth, "admin@"+config["Domain"], []string{to}, []byte("To: "+to+"\nSubject: "+subject+"\nContent-Type: text/html; charset=\"UTF-8\"\n"+msg)); err != nil {
		logger.Println("SendMail:", err.Error())
	}
}

func parseRef(data string) []cid {
	var m = map[cid]struct{}{}
	t := html.NewTokenizer(strings.NewReader(data))
	for {
		t.Next()
		token := t.Token()
		if token.Type == html.ErrorToken {
			break
		}
		if token.Type == html.StartTagToken &&
			token.DataAtom == atom.Blockquote {
			for _, attr := range token.Attr {
				if attr.Key == "cite" {
					if s := attr.Val; strings.HasPrefix(s, "#comment-") {
						id, err := strconv.ParseUint(s[len("#comment-"):], 10, 32)
						if err != nil {
							logger.Println("notification:", err.Error())
							continue
						}
						m[cid(id)] = struct{}{}
					}
					break
				}
			}
		}
	}
	var ret []cid
	for k, _ := range m {
		ret = append(ret, k)
	}
	return ret
}
