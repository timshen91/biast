package main

import (
	"encoding/json"
	"fmt"
	"gedis"
	"strconv"
)

type dbAdapter interface {
	get(id int) (*Article, error)
	set(id int, data *Article)
	keys() []int
}

type redisAdapter struct {
	cli redis.Client
}

func newRedisAdapter(addr, pass, dbId string) (*redisAdapter, error) {
	db, err := strconv.Atoi(dbId)
	if err != nil {
		return nil, err
	}
	return &redisAdapter{
		cli: redis.Client{
			Remote: addr,
			Psw:    pass,
			Db:     db,
		},
	}, nil
}

func (this *redisAdapter) get(id int) (*Article, error) {
	strp, err := this.cli.Get(fmt.Sprint(id))
	if err != nil {
		return nil, err
	}
	str := *strp
	var v Article
	json.Unmarshal([]byte(str), &v)
	return &v, nil
}

func (this *redisAdapter) jsonSet(id int, data string) error {
	err := this.cli.Set(fmt.Sprint(id), data)
	if err != nil {
		return err
	}
	return nil
}

func (this *redisAdapter) set(id int, data *Article) {
	bts, err0 := json.Marshal(data)
	if err0 != nil {
		logger.Println("redisAdapter:", err0.Error())
	}
	err1 := this.jsonSet(id, string(bts))
	if err1 != nil {
		logger.Println("redisAdapter:", err1.Error())
	}
}

func (this *redisAdapter) keys() []int {
	strs, err := this.cli.Keys("") // FIXME what is the fxcking arg?
	if err != nil {
		logger.Println("redisAdapter:", err.Error())
		return nil
	}
	ret := make([]int, 16)
	for _, s := range strs {
		id, err := strconv.Atoi(s)
		if err != nil {
			logger.Println("redisAdapter:", err.Error())
			continue
		}
		ret = append(ret, id)
	}
	return ret
}
