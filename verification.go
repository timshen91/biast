package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func genVerifiCode(w http.ResponseWriter) string {
	rand.Seed(time.Now().Unix())

	numa := rand.Intn(10)
	numb := rand.Intn(10)
	oprand := rand.Intn(3)

	var code string
	var ret int
	switch oprand {
	case 0: // add
		code = strconv.Itoa(numa) + " + " + strconv.Itoa(numb)
		ret = numa + numb

	case 1: // minus
		code = strconv.Itoa(numa) + " - " + strconv.Itoa(numb)
		ret = numa - numb

	case 2: // multiply
		code = strconv.Itoa(numa) + " * " + strconv.Itoa(numb)
		ret = numa * numb
	}
	code = "= " + code

	retstr := strconv.Itoa(ret)
	hret := md5.New()
	io.WriteString(hret, retstr+config["Salt"])
	retstrMd5 := fmt.Sprintf("%x", hret.Sum(nil))

	setCookie("verification", retstrMd5, 3000, w)
	return code
}

func checkVerifiCode(a *http.Request) bool {
	var cret string
	var err error
	if cret, err = getCookie("verification", a); err != nil {
		return false
	}

	ans := strings.TrimSpace(a.Form.Get("verification"))
	hans := md5.New()
	io.WriteString(hans, ans+config["Salt"])
	ansMd5 := fmt.Sprintf("%x", hans.Sum(nil))

	if ansMd5 == cret {
		return true
	}
	return false
}
