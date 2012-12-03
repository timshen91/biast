package main


import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"text/template"
	"time"
)

type Article struct {
	Id         aid
	Author     string
	Email      string
	RemoteAddr string // I'm evil
	Date       time.Time
	Src        string
	Content    string // RAW html
	Notif      bool
	Title      string
	Tags       []string
}

type Comment struct {
	Id         cid
	Author     string
	Email      string
	RemoteAddr string // again, I'm evil
	Date       time.Time
	Website    string
	Content    string // tags limited
	Notif      bool
	Father     aid
}

func (this *Article) getId() uint32 {
	return uint32(this.Id)
}

func (this *Comment) getId() uint32 {
	return uint32(this.Id)
}

func encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, &v)
}

var config = map[string]string{}
var tmpl *template.Template
var logger *log.Logger
var db dbSync

func main() {
	logger.Println("Server start")
	if err := http.ListenAndServe(config["ServerAddr"], nil); err != nil {
		logger.Println("Server cannot start", err.Error())
	}
}

func init() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println(`usage: biast /path/to/config/file`)
		os.Exit(1)
	}
	// config init
	config["Description"] = ""
	buff, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		panic(err.Error())
	}
	for _, line := range strings.Split(string(buff), "\n") {
		if pos := strings.Index(line, "#"); pos != -1 {
			line = line[0:pos]
		}
		if pos := strings.Index(line, "="); pos != -1 {
			config[strings.TrimSpace(line[:pos])] = strings.TrimSpace(line[pos+1:])
		}
	}
	if !checkKeyExist(config, "Domain", "ServerName", "ServerAddr", "DocumentPath", "RootUrl", "AdminList", "DbAddr", "DbPass", "DbId") {
		panic("required config not exist")
	}
	if config["Domain"][len(config["Domain"])-1] == '/' {
		config["Domain"] = config["Domain"][:len(config["Domain"])-1]
	}
	if config["DocumentPath"][len(config["DocumentPath"])-1] != '/' {
		config["DocumentPath"] += "/"
	}
	if config["RootUrl"][0] != '/' {
		config["RootUrl"] = "/" + config["RootUrl"]
	}
	if config["RootUrl"][len(config["RootUrl"])-1] != '/' {
		config["RootUrl"] += "/"
	}
	for _, s := range strings.Split(config["AdminList"], ",") {
		if ss := strings.TrimSpace(s); ss != "" {
			adminList[ss] = struct{}{}
		}
	}
	static := http.FileServer(http.Dir(config["DocumentPath"] + "static/"))
	http.Handle(config["RootUrl"]+"js/", getGzipHandler(handler2HandlerFunc(http.StripPrefix(config["RootUrl"], static))))
	http.Handle(config["RootUrl"]+"css/", getGzipHandler(handler2HandlerFunc(http.StripPrefix(config["RootUrl"], static))))
	http.Handle(config["RootUrl"]+"image/", http.StripPrefix(config["RootUrl"], static))
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
	// module init
	initDb()
	initManager()
	initIndex()
	initPageAdmin()
	initPageArticle()
	initPageAbout()
	initNotification()
	// template init and its coroutine
	go func() {
		for {
			tt := template.New("").Funcs(map[string]interface{}{
				"getCommentList": getCommentList,
			})
			t, err := tt.ParseGlob(config["DocumentPath"] + "template/" + "*")
			if err != nil {
				logger.Println("reparse template failed:", err.Error())
			} else {
				tmpl = t
				updateIndexAndFeed()
			}
			time.Sleep(time.Minute)
		}
	}()
	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch)
		for {
			switch sig := <-ch; sig {
			case os.Interrupt, os.Kill, syscall.SIGTERM: // FIXME: this may not be portable
				logger.Println("Server halt")
				os.Exit(0)
			}
		}
	}()
}

func checkKeyExist(m interface{}, args ...interface{}) bool {
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map || !v.IsValid() {
		return false
	}
	for _, p := range args {
		if !v.MapIndex(reflect.ValueOf(p)).IsValid() {
			return false
		}
	}
	return true
}
