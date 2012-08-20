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
	queueSize     = 16
)

type dbAdapter interface {
	getAll() ([]*Article, []*Comment)
	syncArticle(ptr *Article)
	syncComment(ptr *Comment)
	reset()
}

type redisAdapter struct {
	cli   redis.Client
	queue chan *syncEvent
}

type syncEvent struct {
	id  string
	str string
}

func newRedisAdapter(addr, pass, dbId string) (*redisAdapter, error) {
	db, err := strconv.Atoi(dbId)
	if err != nil {
		return nil, err
	}
	ret := &redisAdapter{
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

func (this *redisAdapter) getAll() ([]*Article, []*Comment) {
	var articleList []*Article
	{
		strs, err := this.cli.Keys(articlePrefix + "*")
		if err != nil {
			logger.Println("redisAdapter:", err.Error())
			return nil, nil
		}
		for _, id := range strs {
			strp, err := this.cli.Get(id)
			if err != nil {
				return nil, nil
			}
			str := *strp // FIXME strp == nil
			var v Article
			json.Unmarshal([]byte(str), &v)
			articleList = append(articleList, &v)
		}
	}
	var commentList []*Comment
	{
		strs, err := this.cli.Keys(commentPrefix + "*")
		if err != nil {
			logger.Println("redisAdapter:", err.Error())
			return nil, nil
		}
		for _, id := range strs {
			strp, err := this.cli.Get(id)
			if err != nil {
				return nil, nil
			}
			str := *strp // FIXME strp == nil
			var v Comment
			json.Unmarshal([]byte(str), &v)
			commentList = append(commentList, &v)
		}
	}
	return articleList, commentList
}

func (this *redisAdapter) syncArticle(article *Article) {
	data := *article
	data.Comments = nil
	bts, err := json.Marshal(data)
	if err != nil {
		logger.Println("redisAdapter:", err.Error())
	}
	this.queue <- &syncEvent{articlePrefix + fmt.Sprint(data.Info.Id), string(bts)}
}

func (this *redisAdapter) syncComment(comment *Comment) {
	bts, err := json.Marshal(comment)
	if err != nil {
		logger.Println("redisAdapter:", err.Error())
	}
	this.queue <- &syncEvent{commentPrefix + fmt.Sprint(comment.Info.Id), string(bts)}
}

func (this *redisAdapter) reset() {
	this.cli.Disconnect()
}

func (this *redisAdapter) sync() {
	for {
		data := <-this.queue
		err := this.cli.Set(data.id, data.str)
		if err != nil {
			logger.Println("redisAdapter:", err.Error())
		}
	}
}
