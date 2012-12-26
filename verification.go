package main

import (
    "math/rand"
    "time"
    "strconv"
    "crypto/md5"
    "io"
    "fmt"
    "sync"
)

type sessionMgr struct {
    lock sync.Mutex
    sessionPool map[string]*sValue
    maxLifeTime int64
}

type sValue struct {
    value string
    time int64
}

var globalSessions *sessionMgr

func genVerifiCode(sid string) (string) {
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

    retstr := strconv.Itoa(ret)
    hret := md5.New()
    io.WriteString(hret, retstr + salt)
    retstrMd5 := fmt.Sprintf("%x", hret.Sum(nil))

    sessionAdd(sid, retstrMd5)
    return code
}

func checkVerifiCode(sid, ans string) (bool) {
    if sv := sessionGet(sid); sv != nil {
        if time.Now().Unix() - sv.time > globalSessions.maxLifeTime {
            sessionRm(sid)
        }

        hans := md5.New()
        io.WriteString(hans, ans + salt)
        ansMd5 := fmt.Sprintf("%x", hans.Sum(nil))
        if ansMd5 == sv.value {
            return true
        }

        sessionRm(sid)
    }
    return false
}

func sessionAdd(sid, retstrMd5 string) {
    globalSessions.lock.Lock()
    defer globalSessions.lock.Unlock()

    sv := &sValue{
        value: retstrMd5,
        time: time.Now().Unix(),
    }
    globalSessions.sessionPool[sid] = sv
}

func sessionGet(sid string) (*sValue) {
    globalSessions.lock.Lock()
    defer globalSessions.lock.Unlock()

    if value, find := globalSessions.sessionPool[sid]; find == true {
        return value
    }
    return nil
}

func sessionRm(sid string) {
    globalSessions.lock.Lock()
    defer globalSessions.lock.Unlock()

    delete(globalSessions.sessionPool, sid)
}

func newSessionMgr() (*sessionMgr) {
    return &sessionMgr{
        sessionPool: make(map[string]*sValue),
        maxLifeTime: 3600,
    }
}

func initSessionMgr() {
    globalSessions = newSessionMgr()
}
