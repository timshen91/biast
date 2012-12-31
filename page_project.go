package main

import (
	"net/http"
)

func projectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	if err := tmpl.ExecuteTemplate(w, "project", map[string]interface{}{
		"config": config,
		"header": "Project",
	}); err != nil {
		logger.Println("project:", err.Error())
	}
}

func initPageProject() {
	http.HandleFunc(config["RootUrl"]+"project", getGzipHandler(projectHandler))
}
