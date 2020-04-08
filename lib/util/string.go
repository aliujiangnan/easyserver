package util

import (
    "unsafe"
    "encoding/json"
)

func BytesToString(b []byte) string {
    return *(*string)(unsafe.Pointer(&b))
}

func StringToBytes(s string) []byte {
    return *(*[]byte)(unsafe.Pointer(&s))
}

func JsonToString(data interface{}) string{
    jso,_ :=json.Marshal(data)
    return string(jso)
}