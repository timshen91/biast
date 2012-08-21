package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"text/template"
	"time"
)

type Article struct {
	Id         aid
	Author     string
	Email      string
	RemoteAddr string // I'm evil
	Date       time.Time

	Title   string
	Content string // RAW html
}

type Comment struct {
	Id         cid
	Author     string
	Email      string
	RemoteAddr string // I'm evil
	Date       time.Time

	Father     aid
	Content    string // plain text
	ReplyNotif bool
}

func (this *Article) getId() uint32 {
	return uint32(this.Id)
}

func (this *Comment) getId() uint32 {
	return uint32(this.Id)
}

var config = map[string]string{}
var tmpl *template.Template
var logger *log.Logger
var artMgr *manager
var db dbSync

func main() {
	// config init
	buff, err := ioutil.ReadFile("/etc/biast.conf")
	if err != nil {
		panic(err.Error())
	}
	for _, line := range strings.Split(string(buff), "\n") {
		if len(line) == 0 {
			continue
		}
		if line[0] == '#' {
			continue
		}
		pos := strings.Index(line, "=")
		if pos != -1 {
			config[strings.TrimSpace(line[:pos])] = strings.TrimSpace(line[pos+1:])
		}
	}
	if !checkKeyExist(config, "ServerName", "ServerAddr", "DocumentPath", "RootUrl", "AdminUrl", "DbAddr", "DbPass", "DbId") {
		panic("config file read failed")
	}
	if config["DocumentPath"][len(config["DocumentPath"])-1] != '/' {
		config["DocumentPath"] += "/"
	}
	if config["RootUrl"][len(config["RootUrl"])-1] != '/' {
		config["RootUrl"] += "/"
	}
	config["TemplatePath"] = config["DocumentPath"] + "template/"
	config["CssPath"] = config["DocumentPath"] + "css/"
	config["CssUrl"] = config["RootUrl"] + "css/"
	config["ImagePath"] = config["DocumentPath"] + "image/"
	config["ImageUrl"] = config["RootUrl"] + "image/"
	http.Handle(config["CssUrl"], http.StripPrefix(config["CssUrl"], http.FileServer(http.Dir(config["CssPath"]))))
	http.Handle(config["ImageUrl"], http.StripPrefix(config["ImageUrl"], http.FileServer(http.Dir(config["ImagePath"]))))
	// template init
	tmpl = template.Must(template.ParseGlob(config["TemplatePath"] + "*.html"))
	var err1 error
	if db, err1 = newRedisSync(config["DbAddr"], config["DbPass"], config["DbId"]); err1 != nil {
		panic(err1.Error())
	}
	// article manager and db init
	artMgr = newArticleMgr(db, runtime.NumCPU())
	// logger
	if _, ok := config["LogPath"]; ok {
		logWriter, err := os.OpenFile(config["LogPath"], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			panic(err.Error())
		}
		logger = log.New(logWriter, "biast: ", log.LstdFlags|log.Lshortfile)
	} else {
		logger = log.New(os.Stderr, "biast: ", log.LstdFlags|log.Lshortfile)
	}

	// modules init
	initPageIndex()
	initPageArticle()
	initPageAdmin()
	initSendMail()
	logger.Println("Server start")
	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch)
		for {
			switch sig := <-ch; sig {
			case os.Interrupt:
				fallthrough
			case os.Kill:
				logger.Println("Server halt")
				os.Exit(0)
			}
		}
	}()
	http.ListenAndServe(config["ServerAddr"], nil)
}

func checkKeyExist(m interface{}, args ...string) bool {
	value := reflect.ValueOf(m)
	if value.Kind() != reflect.Map {
		return false
	}
	tests := make(map[string]struct{})
	for _, s := range args {
		tests[s] = struct{}{}
	}
	keys := value.MapKeys()
	var count int
	for i := range keys {
		_, ok := tests[keys[i].String()]
		if ok {
			count++
		}
	}
	if count == len(args) {
		return true
	}
	return false
}
