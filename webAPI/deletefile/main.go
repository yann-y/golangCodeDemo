package main

import (
	"fmt"
	"os"
)

func main() {

	err := os.MkdirAll("./path", os.ModePerm)
	_, err = os.Create("./path/mp.json")
	if err != nil {
		fmt.Printf("失败！")
	}
}
