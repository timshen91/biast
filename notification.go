package main

import (
	"exp/html"
	"exp/html/atom"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
)

func newCommentNotify(comm *Comment) {
	// notify the article author
	if father := artMgr.getArticle(comm.Father); father.QuoteNotif {
		send(father.Email, "Your article has been commented", fmt.Sprintf(
			`Dear %s, your comment on %s has been commented by %s:
	<blockquote>%s</blockquote>
	Click <a href="%s">here</a> for details
	`, father.Author, config["ServerName"], comm.Author, comm.Content, config["Domain"]+config["ArticleUrl"]+fmt.Sprint(father.Id)+"#comment-"+fmt.Sprint(comm.Id)))
	}
	// notify the commenter
	for _, id := range parseRef(comm.Content) {
		if p := artMgr.getComment(id); p != nil && p.QuoteNotif {
			if comm.Father == p.Father {
				send(p.Email, "Your comment has been quoted", fmt.Sprintf(
					`Dear %s, your comment on %s has been quoted by %s:
<blockquote>%s</blockquote>
Click <a href="%s">here</a> for details
`, p.Author, config["ServerName"], comm.Author, comm.Content, config["Domain"]+config["ArticleUrl"]+fmt.Sprint(p.Father)+"#comment-"+fmt.Sprint(comm.Id)))
			}
		}
	}
}

var mailAuth smtp.Auth

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
}

func send(to, subject, msg string) {
	logger.Println("notification:", "SendMail:", msg, "for", to)
	if err := smtp.SendMail(config["SMTPAddr"], mailAuth, "admin@"+config["Domain"], []string{to}, []byte("To: "+to+"\nSubject: "+subject+"\nContent-Type: text/html; charset=\"UTF-8\"\n"+msg)); err != nil {
		logger.Println("SendMail:", err.Error())
	}
}

func parseRef(data string) []cid {
	var ret []cid
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
						ret = append(ret, cid(id))
					}
					break
				}
			}
		}
	}
	return ret
}
