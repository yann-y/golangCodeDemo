package main

import (
	"fmt"
	"strconv"
	"time"
)

func main() {

	tstring := "1632794347377"
	//timestamp check  五分钟的时效性
	ts, err := strconv.Atoi(tstring)
	if err != nil {
		fmt.Println("timestamp error: %s", err.Error())
	}
	now := time.Now()

	paramTime := time.Unix(0, int64(ts)*int64(time.Millisecond))
	fmt.Println(paramTime.String())
	paramTime = paramTime.Add(time.Minute * 15)
	fmt.Println(paramTime.String())
	if now.After(paramTime) {
		fmt.Println("request expired")
	} else {
		fmt.Println("未过期")
	}

	//获取当前时间
	newNow := time.Now()
	newNow = newNow.Add(time.Hour * 2)
	fmt.Println(newNow.String())
	fmt.Println(newNow.Unix())
}
