package main

import (
	"encoding/json"
	"fmt"
	"goredis"
	"strconv"
)

const (
	articlePrefix = "article"
	queueSize = 16
)

type dbAdapter interface {
	getArticle(id uint32) (*Article, error)
	setArticle(ptr *Article)
	articleKeys() []uint32
	reset()
}

type redisAdapter struct {
	cli redis.Client
	queue chan *Article
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
		queue: make(chan *Article, queueSize),
	}
	if err := ret.cli.Connect(); err != nil {
		return nil, err
	}
	go ret.sync()
	return ret, nil
}

func (this *redisAdapter) getArticle(id uint32) (*Article, error) {
	strp, err := this.cli.Get(articlePrefix + fmt.Sprint(id))
	if err != nil {
		return nil, err
	}
	str := *strp // FIXME strp == nil
	var v Article
	json.Unmarshal([]byte(str), &v)
	return &v, nil
}

func (this *redisAdapter) setArticle(data *Article) {
	this.queue<-data
}

func (this *redisAdapter) articleKeys() []uint32 {
	strs, err := this.cli.Keys(articlePrefix + "*")
	if err != nil {
		logger.Println("redisAdapter:", err.Error())
		return nil
	}
	ret := make([]uint32, 0)
	for _, s := range strs {
		id, err := strconv.ParseUint(s[len(articlePrefix):], 10, 32)
		if err != nil {
			logger.Println("redisAdapter:", err.Error())
			continue
		}
		ret = append(ret, uint32(id))
	}
	return ret
}

func (this *redisAdapter) reset() {
	this.cli.Disconnect()
}

func (this *redisAdapter) sync() {
	for {
		article := <-this.queue
		bts, err0 := json.Marshal(article)
		if err0 != nil {
			logger.Println("redisAdapter:", err0.Error())
		}
		err1 := this.cli.Set(articlePrefix + fmt.Sprint(article.Id), string(bts))
		if err1 != nil {
			logger.Println("redisAdapter:", err1.Error())
		}
	}
}
