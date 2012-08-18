package main

import (
	"os"
	"io/ioutil"
	"net/http"
	"text/template"
	"strings"
	"log"
	"time"
	"reflect"
)

type info struct {
    author string
    email string
    content string // plain html
    date time.Time
}

type article struct {
	info
    id int
    title string
    comments []*comment
}

type comment struct {
	info
	father int
}

var articles []*article
var config map[string]string = make(map[string]string)
var tmpl *template.Template
var db dbAdapter
var logger *log.Logger

func dbReadin() {
	idList := db.keys()
	var max int
	for _, id := range idList {
		if max < id {
			max = id
		}
	}
	articles = make([]*article, max)
	for _, id := range idList {
		p, err := db.get(id)
		if err != nil {
			logger.Println(err.Error())
		}
		articles[id] = p.(*article)
	}
}

func saveCron() {
	// TODO
}

func checkKeyExist(m interface{}, args ...string) bool {
	value := reflect.ValueOf(m)
	if value.Kind() != reflect.Map {
		return false
	}
	tests := make(map[string]bool)
	for _, s := range args {
		tests[s] = true
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

func main() {
	buff, err0 := ioutil.ReadFile("/etc/biast.conf")
	if err0 != nil {
		panic(err0.Error())
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
			config[strings.TrimSpace(line[:pos])] = strings.TrimSpace(line[pos + 1:])
		}
	}
	if !checkKeyExist(config, "ServerName", "ServerAddr", "DocumentPath", "RootUrl", "DbAddr", "DbPass", "DbId") {
		panic("config file read failed")
	}
	if config["DocumentPath"][len(config["DocumentPath"]) - 1] != '/' {
		config["DocumentPath"] += "/"
	}
	if config["RootUrl"][len(config["RootUrl"]) - 1] != '/' {
		config["RootUrl"] += "/"
	}
	config["TemplatePath"] = config["DocumentPath"]+ "template/"
	config["CssPath"] = config["DocumentPath"] + "css/"
	config["CssUrl"] = config["RootUrl"] + "css/"
	config["ImagePath"] = config["DocumentPath"] + "image/"
	config["ImageUrl"] = config["RootUrl"] + "image/"
	if _, ok := config["LogPath"]; !ok {
		config["LogPath"] = config["documentPath"] + "log"
	}
	tmpl = template.Must(template.ParseGlob(config["TemplatePath"] + "*"))
	db = &redisAdapter {
		config["DbAddr"],
		config["DbPass"],
		config["DbId"],
	}
	logWriter, err := os.OpenFile(config["LogPath"], os.O_APPEND | os.O_CREATE, 0644)
	if err != nil {
		panic(err.Error())
	}
	logger = log.New(logWriter, "biast: ", log.LstdFlags | log.Lshortfile)
	http.Handle(config["CssUrl"], http.StripPrefix(config["CssUrl"], http.FileServer(http.Dir(config["CssPath"]))))
	http.Handle(config["ImageUrl"], http.StripPrefix(config["ImageUrl"], http.FileServer(http.Dir(config["ImagePath"]))))

	dbReadin()
	// modules init
	// indexInit()
	// articleInit()
	// postArticleInit()
	go saveCron()
	http.ListenAndServe(config["ServerAddr"], nil)
}
