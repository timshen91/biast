package main

import (
	"encoding/json"
	"fmt"
	"goredis"
	"strconv"
)

const (
	articlePrefix = "article"
	commentPrefix = "comment"
	notiPrefix    = "comment"
	queueSize     = 16
)

type dbSync interface {
	getArticles() []*Article
	syncArticle(*Article)
	syncComment(*Comment)
	getComments() []*Comment
	syncNotiInfo(*Noti)
	getNotiInfo() []*Noti
	reset()
}

type redisSync struct {
	cli   redis.Client
	queue chan *syncEvent
}

type syncEvent struct {
	id  string
	str string
}

func newRedisSync(addr, pass, dbId string) (*redisSync, error) {
	db, err := strconv.Atoi(dbId)
	if err != nil {
		return nil, err
	}
	ret := &redisSync{
		cli: redis.Client{
			Remote: addr,
			Psw:    pass,
			Db:     db,
		},
		queue: make(chan *syncEvent, queueSize),
	}
	if err := ret.cli.Connect(); err != nil {
		return nil, err
	}
	go ret.sync()
	return ret, nil
}

func (this *redisSync) getArticles() []*Article {
	var ret []*Article
	var temp *Article
	this.getStrList(articlePrefix, func(str []byte) {
		temp = nil
		json.Unmarshal(str, &temp)
		ret = append(ret, temp)
	})
	return ret
}

func (this *redisSync) syncArticle(article *Article) {
	data := *article
	data.Comments = nil
	bts, err := json.Marshal(data)
	if err != nil {
		logger.Println("redisSync:", err.Error())
	}
	this.queue <- &syncEvent{articlePrefix + fmt.Sprint(data.Info.Id), string(bts)}
}

func (this *redisSync) getComments() []*Comment {
	var ret []*Comment
	var temp *Comment
	this.getStrList(commentPrefix, func(str []byte) {
		temp = nil
		json.Unmarshal(str, &temp)
		ret = append(ret, temp)
	})
	return ret
}

func (this *redisSync) syncComment(comment *Comment) {
	bts, err := json.Marshal(comment)
	if err != nil {
		logger.Println("redisSync:", err.Error())
	}
	this.queue <- &syncEvent{commentPrefix + fmt.Sprint(comment.Info.Id), string(bts)}
}

func (this *redisSync) getNotiInfo() []*Noti {
	var ret []*Noti
	var temp *Noti
	this.getStrList(notiPrefix, func(str []byte) {
		temp = nil
		json.Unmarshal(str, &temp)
		ret = append(ret, temp)
	})
	return ret
}

func (this *redisSync) syncNotiInfo(p *Noti) {
}

func (this *redisSync) reset() {
	this.cli.Disconnect()
}

func (this *redisSync) getStrList(prefix string, callback func([]byte)) {
	strs, err := this.cli.Keys(prefix + "*")
	if err != nil {
		logger.Println("redisSync:", err.Error())
		return
	}
	for _, id := range strs {
		str, err := this.cli.Get(id)
		if err != nil {
			logger.Println("redisSync:", err.Error())
			continue
		}
		// v = nil
		// json.Unmarshal([]byte(str), &v)
		// list = append(list, &v)
		callback([]byte(str))
	}
	return
}

func (this *redisSync) sync() {
	for {
		data := <-this.queue
		err := this.cli.Set(data.id, data.str)
		println(data.id, data.str)
		if err != nil {
			logger.Println("redisSync:", err.Error())
		}
	}
}
