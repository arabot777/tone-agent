package reader

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type Reader struct{}

func New() *Reader {
	return &Reader{}
}

// BoolVar 解析布尔类型的环境变量
func (e *Reader) BoolVar(p *bool, key string, val bool) {
	v, ok := os.LookupEnv(key)
	if !ok {
		*p = val
		return
	}
	if v != "true" && v != "false" && v != "1" && v != "0" {
		fmt.Printf("warning: os env [%s] should be a bool(true/false/1/0), but got: [%s], use default value\n", key, v)
		*p = val
		return
	}
	if v == "true" || v == "1" {
		*p = true
	} else {
		*p = false
	}
}

// IntVar 解析数字类型的环境变量
func (e *Reader) IntVar(p *int, key string, val int) {
	v, ok := os.LookupEnv(key)
	if !ok {
		*p = val
		return
	}
	v_, err := strconv.Atoi(v)
	if err != nil {
		fmt.Printf("warning: os env [%s] should be an integer, but got: [%s], use default value\n", key, v)
		*p = val
		return
	}
	*p = v_
}

// StringVar 解析字符串类型的环境变量
func (e *Reader) StringVar(p *string, key string, val string) {
	v, ok := os.LookupEnv(key)
	if !ok {
		*p = val
		return
	}
	*p = v
}

// MapVar 解析map类型的环境变量
func (e *Reader) MapVar(p *map[string]string, key string, val map[string]string) {
	v, ok := os.LookupEnv(key)
	if !ok {
		*p = val
		return
	}
	err := json.Unmarshal([]byte(v), p)
	if err != nil {
		fmt.Printf("warning: os env [%s] should be a map, but got: [%s], use default value\n", key, v)
		*p = val
		return
	}
}
