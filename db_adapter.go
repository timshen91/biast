package main

import (
	"fmt"
	"gedis"
	"strconv"
	"encoding/json"
)

type dbAdapter interface {
	get(id int) (*article, error)
	set(id int, data *article) error
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
		cli : redis.Client{
			Remote : addr,
			Psw : pass,
			Db : db,
		},
	}, nil
}

func (this *redisAdapter) get(id int) (*article, error) {
	strp, err := this.cli.Get(fmt.Sprint(id))
	if err != nil {
		return nil, err
	}
	str := *strp
	var v article
	json.Unmarshal([]byte(str), &v)
	return &v, nil
}

func (this *redisAdapter) set(id int, data *article) error {
	bts, err0 := json.Marshal(data)
	if err0 != nil {
		return err0
	}
	err1 := this.cli.Set(fmt.Sprint(id), string(bts))
	if err1 != nil {
		return err1
	}
	return nil
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
