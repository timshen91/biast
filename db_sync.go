package main

import (
	"encoding/json"
	"fmt"
	"goredis"
	"strconv"
)

type prefixType string

type hasId interface {
	getId() uint32
}

const (
	articlePrefix prefixType = "article"
	commentPrefix            = "comment"
	queueSize                = 16
)

type dbSync interface {
	getStrList(prefixType) [][]byte
	sync(prefixType, hasId)
}

type redisSync struct {
	cli   redis.Client
	queue chan *syncReq
}

type syncReq struct {
	id  []byte
	str []byte
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
		queue: make(chan *syncReq, queueSize),
	}
	if err := ret.cli.Connect(); err != nil {
		return nil, err
	}
	go ret.worker()
	return ret, nil
}

func (this *redisSync) getStrList(prefix prefixType) [][]byte {
	strs, err := this.cli.Keys([]byte(prefix + "*"))
	if err != nil {
		logger.Println("redisSync:", err.Error())
		return nil
	}
	var ret [][]byte
	for _, id := range strs {
		str, err := this.cli.Get(id)
		if err != nil {
			logger.Println("redisSync:", err.Error())
			continue
		}
		ret = append(ret, str)
	}
	return ret
}

func (this *redisSync) sync(prefix prefixType, data hasId) {
	bts, err := json.Marshal(data)
	if err != nil {
		logger.Println("redisSync:", err.Error())
	}
	this.queue <- &syncReq{[]byte(string(prefix) + fmt.Sprint(data.getId())), bts}
}

func (this *redisSync) worker() {
	for {
		data := <-this.queue
		err := this.cli.Set(data.id, data.str)
		if err != nil {
			logger.Println("redisSync:", err.Error())
		}
	}
}
