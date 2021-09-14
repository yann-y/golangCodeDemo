package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {

	fmt.Printf(strMd5("C:\\Users\\严卓泉\\Downloads\\ubuntu-20.04.2.0-desktop-amd64.iso"))
}
func strMd5(str string) (retMd5 string) {
	f, err := os.Open(str)
	if err != nil {
		return "md5 Open file err"
	}
	defer f.Close()
	body, err := ioutil.ReadAll(f)
	if err != nil {
		return "ioutil.ReadAll"
	}
	return fmt.Sprintf("%x", md5.Sum(body))
}
