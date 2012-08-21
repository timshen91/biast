package main

import (
	"encoding/json"
	"sync/atomic"
)

type aid uint32
type cid uint32

type manager struct { // assume pointer assignment is atomic
	articles     map[aid]*Article
	comments     map[cid]*Comment
	commentLists map[aid][]*Comment
	articleHead  aid
	commentHead  cid
}

func newArticleMgr(db dbSync, poolSize int) *manager {
	ret := &manager{
		articles:     map[aid]*Article{},
		comments:     map[cid]*Comment{},
		commentLists: map[aid][]*Comment{},
		articleHead:  0,
		commentHead:  0,
	}
	for _, bts := range db.getStrList(articlePrefix) {
		var p *Article
		json.Unmarshal(bts, &p)
		ret.articles[p.Id] = p
		if ret.articleHead < p.Id {
			ret.articleHead = p.Id
		}
	}
	for _, bts := range db.getStrList(commentPrefix) {
		var p *Comment
		json.Unmarshal(bts, &p)
		ret.comments[p.Id] = p
		ret.commentLists[p.Father] = append(ret.commentLists[p.Father], p)
		if ret.commentHead < p.Id {
			ret.commentHead = p.Id
		}
	}
	for _, p := range ret.commentLists {
		qsortForCommentList(p, 0, len(p)-1)
	}
	return ret
}

func (this *manager) atomGetArticle(id aid) *Article {
	ret, ok := this.articles[id]
	if !ok {
		return nil
	}
	return ret
}

func (this *manager) atomSetArticle(p *Article) {
	id := p.Id
	this.articles[id] = p
}

func (this *manager) atomGetComment(id cid) *Comment {
	ret, ok := this.comments[id]
	if !ok {
		return nil
	}
	return ret
}

func (this *manager) atomGetCommentList(id aid) []*Comment {
	ret, ok := this.commentLists[id]
	if !ok {
		return nil
	}
	return ret
}

func (this *manager) atomAppendComment(p *Comment) { // FIXME probably not safe
	id := p.Father
	this.commentLists[id] = append(this.commentLists[id], p)
}

func (this *manager) atomGetAllArticles() []*Article {
	ret := make([]*Article, 0)
	for _, p := range this.articles {
		ret = append(ret, p)
	}
	return ret
}

func (this *manager) allocArticleId() aid {
	return aid(allocId((*uint32)(&(this.articleHead))))
}

func (this *manager) allocCommentId() cid {
	return cid(allocId((*uint32)(&this.commentHead)))
}

func allocId(head *uint32) uint32 {
	return atomic.AddUint32(head, 1)
}

func qsortForCommentList(a []*Comment, l, r int) {
	if l > r {
		return
	}
	i := l
	j := (r-l)/2 + l
	a[i], a[j] = a[j], a[i]
	j = l
	for i = l + 1; i <= r; i++ {
		if a[i].Id < a[l].Id {
			j++
			a[j], a[i] = a[i], a[j]
		}
	}
	a[j], a[l] = a[l], a[j]
	qsortForCommentList(a, l, j-1)
	qsortForCommentList(a, j+1, r)
}
