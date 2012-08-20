package main

import (
	"sort"
	"sync"
	"sync/atomic"
)

type manager struct {
	articles    map[uint32]*Article
	mutex       sync.RWMutex
	articleHead uint32
	commentHead uint32
}

func newArticleMgr(db dbAdapter) *manager {
	ret := &manager{
		articles: map[uint32]*Article{},
	}
	artList, commList := db.getAll()
	ret.articleHead = 0
	for _, p := range artList {
		ret.articles[p.Info.Id] = p
		p.Comments = make([]*Comment, 0)
		if ret.articleHead < p.Info.Id {
			ret.articleHead = p.Info.Id
		}
	}
	ret.commentHead = 0
	for _, p := range commList {
		f, ok := ret.articles[p.Father]
		if !ok {
			logger.Println("comment without a father:", p)
			continue
		}
		f.Comments = append(f.Comments, p)
		if ret.articleHead < p.Info.Id {
			ret.articleHead = p.Info.Id
		}
	}
	for _, p := range artList {
		var temp commentList = p.Comments
		sort.Sort(temp)
	}
	return ret
}

func (this *manager) atomGet(id uint32) *Article {
	this.mutex.RLock()
	ret, ok := this.articles[id]
	this.mutex.RUnlock()
	if !ok {
		return nil
	}
	return ret
}

func (this *manager) atomSet(ptr *Article) {
	this.mutex.Lock()
	this.articles[ptr.Info.Id] = ptr
	this.mutex.Unlock()
}

func (this *manager) values() []*Article {
	ret := make([]*Article, 0)
	this.mutex.RLock()
	for _, p := range this.articles {
		ret = append(ret, p)
	}
	this.mutex.RUnlock()
	return ret
}

func allocId(head *uint32) uint32 {
	return atomic.AddUint32(head, 1)
}

func (this *manager) allocArticleId() uint32 {
	return allocId(&this.articleHead)
}

func (this *manager) allocCommentId() uint32 {
	return allocId(&this.commentHead)
}

type commentList []*Comment

func (this commentList) Len() int {
	return len(this)
}

func (this commentList) Less(i, j int) bool {
	return this[i].Info.Date.Unix() < this[j].Info.Date.Unix()
}

func (this commentList) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}
