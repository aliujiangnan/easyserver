package util

import (
	"time"
	"fmt"
	"strconv"
    "crypto/md5"
)

type Token struct{
	id	int
	time int64
	period int64
}

var gTokens = make(map[string]*Token)
var gUsers = make(map[int]string)

func Sign(str string) string {
    data := []byte(str)
    has := md5.Sum(data)
    md5str := fmt.Sprintf("%x", has)
    return md5str
}

func NewToken(id int, period int) string{
	token := gUsers[id]
	if token != "" {
		DelToken(token)
	}

	now := time.Now().Unix()
	tokenStr := Sign(strconv.Itoa(id) + "!@#$%^&" + strconv.Itoa(int(now)))
	gTokens[tokenStr] = &Token{id, now, int64(period)}
	gUsers[id] = token
	return token
}

func GetToken(id int)string{
	return gUsers[id]
}

func GetUserID(token string) int{
	return gTokens[token].id
}

func IsValid(token string) bool{
	info := gTokens[token]
	if info == nil {
		return false
	}
	if info.time + info.period < time.Now().Unix() {
		return false
	}
	return true
}

func DelToken(token string){
	info := gTokens[token]
	if info != nil {
		delete(gTokens, token)
		delete(gUsers, info.id)
	}
}