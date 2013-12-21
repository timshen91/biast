package main

import (
	"fmt"
	"github.com/novanchao/redisgo"
	"strconv"
)

type prefixType string

const (
	articlePrefix prefixType = "a:"
	commentPrefix            = "c:"
	queueSize                = 16
)

type dbSync interface {
	getStrList(prefixType) ([]string, [][]byte)
	sync(prefixType, string, interface{})
}

type redisSync struct {
	cli   redis.Client
	queue chan *syncReq
}

type syncReq struct {
	prefix prefixType
	id     string
	data   interface{}
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
	go func() {
		for {
			req := <-ret.queue
			bts, err := encode(req.data)
			if err != nil {
				logger.Println("redisSync:", err.Error())
				continue
			}
			if err := ret.cli.Set([]byte(string(req.prefix)+fmt.Sprint(req.id)), bts); err != nil {
				logger.Println("redisSync:", err.Error())
			}
		}
	}()
	return ret, nil
}

func (this *redisSync) getStrList(prefix prefixType) (idList []string, valueList [][]byte) {
	strs, err := this.cli.Keys([]byte(prefix + "*"))
	if err != nil {
		logger.Println("redisSync:", err.Error())
		return nil, nil
	}
	var ret [][]byte
	for _, id := range strs {
		str, err := this.cli.Get(id)
		if err != nil {
			logger.Println("redisSync:", err.Error())
			continue
		}
		idList = append(idList, string(id))
		ret = append(ret, str)
	}
	return idList, ret
}

func (this *redisSync) sync(prefix prefixType, id string, data interface{}) {
	this.queue <- &syncReq{prefix, id, data}
}

func initDb() {
	var err error
	if db, err = newRedisSync(config["DbAddr"], config["DbPass"], config["DbId"]); err != nil {
		panic(err.Error())
	}
}
