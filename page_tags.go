package main

import (
	"net/http"
	"sort"
)

var tags2Article = map[string]map[aid]struct{}{}

func tagHandler(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Path[len(config["TagsUrl"]):]
	if indexList := getArticleByTag(tag); indexList == nil {
		if err := tmpl.ExecuteTemplate(w, "tags", map[string]interface{}{
			"config": config,
			"tags":   getAllTags(),
			"header": "Tags Cloud",
		}); err != nil {
			logger.Println("tags:", err.Error())
		}
	} else {
		if err := tmpl.ExecuteTemplate(w, "index", map[string]interface{}{
			"config":   config,
			"articles": indexList,
			"header":   "Tag: " + tag,
		}); err != nil {
			logger.Println("tags:", err.Error())
		}
	}
}

func getAllTags() []string {
	ret := []string{}
	for k, _ := range tags2Article {
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return ret
}

func getArticleByTag(t string) []*Article {
	var ret []*Article
	for id, _ := range tags2Article[t] {
		ret = append(ret, getArticle(id))
	}
	qsortForArticleList(ret, 0, len(ret)-1)
	return ret
}

func updateTags(id aid, old, tags []string) {
	for _, tag := range old {
		delete(tags2Article[tag], id)
	}
	for _, tag := range tags {
		if _, ex := tags2Article[tag]; !ex {
			tags2Article[tag] = map[aid]struct{}{}
		}
		tags2Article[tag][id] = struct{}{}
	}
	for _, tag := range old {
		if len(tags2Article[tag]) == 0 {
			delete(tags2Article, tag)
		}
	}
}

func initTags() {
	for _, article := range getArticleList() {
		id := article.Id
		for _, tag := range article.Tags {
			if _, ex := tags2Article[tag]; !ex {
				tags2Article[tag] = map[aid]struct{}{}
			}
			tags2Article[tag][id] = struct{}{}
		}
	}
}

func init() {
	config["TagsUrl"] = config["RootUrl"] + "tags/"
	http.HandleFunc(config["TagsUrl"], tagHandler)
}
