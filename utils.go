package main

import (
	"errors"
	"net/http"
	"net/url"
)

func setCookie(key, value string, maxage int, w http.ResponseWriter) {
    c := &http.Cookie{
        Name:   key,
        Value:  url.QueryEscape(value),
        MaxAge: maxage,
    }
    http.SetCookie(w, c)
}

func getCookie(key string, r *http.Request) (string, error) {
    if c, err := r.Cookie(key); err == nil {
        if ret, err1 := url.QueryUnescape(c.Value); err1 == nil {
            return ret, nil
        }
    }
    return "", errors.New("invalid url escaped cookie")
}
