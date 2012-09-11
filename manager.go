package main

import (
	"fmt"
	"sync/atomic"
)

type aid uint32
type cid uint32

// assume pointer assignment is atomic
var articles = map[aid]*Article{}
var comments = map[cid]*Comment{}
var commentLists = map[aid][]*Comment{}
var adminList = map[string]struct{}{}
var articleHead aid
var commentHead cid

func getArticleList() []*Article {
	ret := make([]*Article, 0)
	for _, p := range articles {
		ret = append(ret, p)
	}
	return ret
}

func getCommentList(id aid) []*Comment {
	ret, ok := commentLists[id]
	if !ok {
		return nil
	}
	return ret
}

func getArticle(id aid) *Article {
	ret, ok := articles[id]
	if !ok {
		return nil
	}
	return ret
}

func setArticle(p *Article) {
	id := p.Id
	articles[id] = p
	db.sync(articlePrefix, fmt.Sprint(p.Id), p)
}

func getComment(id cid) *Comment {
	ret, ok := comments[id]
	if !ok {
		return nil
	}
	return ret
}

func setComment(c *Comment) {
	id := c.Id
	comments[id] = c
	db.sync(commentPrefix, fmt.Sprint(c.Id), c)
	commentList := commentLists[c.Father]
	for i, _ := range commentList {
		if commentList[i].Id == id {
			commentList[i] = c
			break
		}
	}
}

func appendComment(p *Comment) { // FIXME probably not safe
	id := p.Father
	commentLists[id] = append(commentLists[id], p)
	db.sync(commentPrefix, fmt.Sprint(p.Id), p)
}

func allocArticleId() aid {
	return aid(allocId((*uint32)(&(articleHead))))
}

func allocCommentId() cid {
	return cid(allocId((*uint32)(&commentHead)))
}

func allocId(head *uint32) uint32 {
	return atomic.AddUint32(head, 1)
}

func initManager() {
	initArticleList()
	initCommentList()
	initPageTags()
}

func initArticleList() {
	_, vList := db.getStrList(articlePrefix)
	for _, bts := range vList {
		var p *Article
		if err := decode(bts, &p); err != nil {
			logger.Println("manager:", err.Error())
			continue
		}
		articles[p.Id] = p
		if p.Src == "" {
			p.Src = p.Content
		}
		if articleHead < p.Id {
			articleHead = p.Id
		}
	}
}

func initCommentList() {
	_, vList := db.getStrList(commentPrefix)
	for _, bts := range vList {
		var p *Comment
		if err := decode(bts, &p); err != nil {
			logger.Println("manager:", err.Error())
			continue
		}
		comments[p.Id] = p
		commentLists[p.Father] = append(commentLists[p.Father], p)
		if commentHead < p.Id {
			commentHead = p.Id
		}
	}
	for _, p := range commentLists {
		sortSlice(p, func(a, b interface{}) bool {
			return a.(*Comment).Id < b.(*Comment).Id
		})
	}
}
