package main

import (
    "math/rand"
    "time"
    "strconv"
    "crypto/md5"
    "io"
    "fmt"
    "net/http"
    "strings"
)

func genVerifiCode(w http.ResponseWriter) (string) {
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

        case 2: // plus
        code = strconv.Itoa(numa) + " * " + strconv.Itoa(numb)
        ret = numa * numb
    }
    code = "= " + code

    retstr := strconv.Itoa(ret)
    hret := md5.New()
    io.WriteString(hret, retstr + salt)
    retstrMd5 := fmt.Sprintf("%x", hret.Sum(nil))

    setCookie("vcode", retstrMd5, config["RootUrl"], config["Domain"], 300, w)
    return code
}

func checkVerifiCode(a *http.Request) (bool) {
    var cret string
    var err error
    if cret, err = getCookie("vcode", a); err != nil {
        return false
    }

    ans := strings.TrimSpace(a.Form.Get("verification"))
    hans := md5.New()
    io.WriteString(hans, ans + salt)
    ansMd5 := fmt.Sprintf("%x", hans.Sum(nil))

    if ansMd5 == cret {
        return true
    }
    return false
}
