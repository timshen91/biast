package main

import (
	"net/smtp"
)

func newCommentNotify(comm *Comment) {
	for _, id := range parseRef(comm.Content) {
		if p := artMgr.atomGetComment(id); p != nil && p.ReplyNotif {
			send(p.Email, "notification\n")
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
	if _, ok := config["SMTPFrom"]; !ok {
		config["SMTPFrom"] = config["ServerName"]
	}
	mailAuth = smtp.PlainAuth("", config["SMTPUsername"], config["SMTPPass"], config["SMTPAddr"])
}

func send(to string, msg string) {
	if err := smtp.SendMail(config["SMTPAddr"], mailAuth, config["SMTPFrom"], []string{to}, []byte(msg)); err != nil {
		logger.Println("SendMail:", err.Error())
	}
}

func parseRef(data string) []cid { // TODO
	return nil
}
