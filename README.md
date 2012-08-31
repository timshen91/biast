## A blog system

It's just a practice for our web application experience. We apply exact evil premature optimazation on it.

## Written in Go, intended to be runtime efficient

[Go](http://golang.org) is a young but well-designed language. There's not many new concepts in it, but the best.

## Dependency

* [Go hg](http://https://code.google.com/p/go/)

	You need to compile go source to get "exp" package which we've used("exp/html" and "exp/html/atom")

* [Redis](http://redis.io)

	Yes we do use a database, because we don't think raw file is a good choice for data persistence and update. And Redis is fast and simple.
