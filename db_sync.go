package main

import (
	"encoding/json"
	"fmt"
	"github.com/nokivan/redis.go"
	"strconv"
)

const (
	articlePrefix prefixType = "article"
	commentPrefix            = "comment"
	queueSize                = 16
)

type prefixType string

type hasId interface {
	getId() uint32
}

type dbSync interface {
	getStrList(prefixType) [][]byte
	sync(prefixType, hasId)
}

type redisSync struct {
	cli   redis.Client
	queue chan *syncReq
}

type syncReq struct {
	prefix prefixType
	data   hasId
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
	this.queue <- &syncReq{prefix, data}
}

func (this *redisSync) worker() {
	for {
		req := <-this.queue
		bts, err := json.Marshal(req.data)
		if err != nil {
			logger.Println("redisSync:", err.Error())
			continue
		}
		if err := this.cli.Set([]byte(string(req.prefix)+fmt.Sprint(req.data.getId())), bts); err != nil {
			logger.Println("redisSync:", err.Error())
		}
	}
}
