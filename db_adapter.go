package main

type serializable interface {
	serialize() string
}

type dbAdapter interface {
	get(id int) (ptr interface{}, err error)
	set(id int, data string)
	keys() []int
}

type redisAdapter struct {
	addr string
	pass string
	dbId string
}

func (this *redisAdapter) get(id int) (ptr interface{}, err error) {
	return nil, nil
}

func (this *redisAdapter) set(id int, data string) {
}

func (this *redisAdapter) keys() []int {
	return nil
}
