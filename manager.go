package main

import (
	"sort"
	"sync"
	"sync/atomic"
)

type commentList []*Comment

type manager struct {
	mutex       []sync.RWMutex
	articles    map[uint32]*Article
	articleHead uint32
	commentHead uint32
}

func newArticleMgr(db dbSync, poolSize int) *manager {
	ret := &manager{
		articles:    map[uint32]*Article{},
		articleHead: 0,
		commentHead: 0,
		mutex:       make([]sync.RWMutex, poolSize),
	}
	artList := db.getArticles()
	for _, p := range artList {
		ret.articles[p.Info.Id] = p
		p.Comments = make([]*Comment, 0)
		if ret.articleHead < p.Info.Id {
			ret.articleHead = p.Info.Id
		}
	}
	commList := db.getComments()
	for _, p := range commList {
		f, ok := ret.articles[p.Father]
		if !ok {
			logger.Println("comment without a father:", p)
			continue
		}
		f.Comments = append(f.Comments, p)
		if ret.commentHead < p.Info.Id {
			ret.commentHead = p.Info.Id
		}
	}
	for _, p := range artList {
		var temp commentList = p.Comments
		sort.Sort(temp)
	}
	return ret
}

func (this *manager) atomGetArticle(id uint32) *Article {
	this.mutex[id%uint32(len(this.mutex))].RLock()
	ret, ok := this.articles[id]
	this.mutex[id%uint32(len(this.mutex))].RUnlock()
	if !ok {
		return nil
	}
	return ret
}

func (this *manager) atomSetArticle(p *Article) {
	id := p.Info.Id
	this.mutex[id%uint32(len(this.mutex))].Lock()
	this.articles[id] = p
	this.mutex[id%uint32(len(this.mutex))].Unlock()
}

func (this *manager) atomAppendComment(p *Comment) {
	id := p.Father
	this.mutex[id%uint32(len(this.mutex))].Lock()
	art := this.articles[id]
	art.Comments = append(art.Comments, p)
	this.mutex[id%uint32(len(this.mutex))].Unlock()
}

func (this *manager) atomGetAllArticles() []*Article {
	ret := make([]*Article, 0)
	for _, p := range this.mutex {
		p.RLock()
	}
	for _, p := range this.articles {
		ret = append(ret, p)
	}
	for _, p := range this.mutex {
		p.RUnlock()
	}
	return ret
}

func (this *manager) allocArticleId() uint32 {
	return allocId(&this.articleHead)
}

func (this *manager) allocCommentId() uint32 {
	return allocId(&this.commentHead)
}

func allocId(head *uint32) uint32 {
	return atomic.AddUint32(head, 1)
}

func (this commentList) Len() int {
	return len(this)
}

func (this commentList) Less(i, j int) bool {
	return this[i].Info.Date.Unix() < this[j].Info.Date.Unix()
}

func (this commentList) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}
