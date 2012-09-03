package main

import (
	"net/http"
)

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	if err := tmpl.ExecuteTemplate(w, "about", map[string]interface{}{
		"config": config,
		"header": "About",
	}); err != nil {
		logger.Println("about:", err.Error())
	}
}

func initPageAbout() {
	http.HandleFunc(config["RootUrl"]+"about", getGzipHandler(aboutHandler))
}
