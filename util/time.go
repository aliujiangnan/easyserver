package util

import (
	"time"
)

func SetTimeout(cb func(), d int) {
	go func(){
		time.Sleep(time.Second * time.Duration(d))
		cb()
	}()
}