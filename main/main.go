package main

import (
	"crypto/md5"

	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	//596e52b5c6ae874f391e8ee73247ea46
	//1b17240ad6eb6f3b69f64a321ddf4fb5
	//3ed3edebda016c48925c569d17aa53c3
	fmt.Println(strMd5("C:\\Users\\严卓泉\\Downloads\\wps_symbol_fonts.zip"))
}
func strMd5(str string) (retMd5 string) {
	f, err := os.Open(str)
	if err != nil {
		str1 := "Open err"
		return str1
	}
	defer f.Close()

	body, err := ioutil.ReadAll(f)
	if err != nil {
		str2 := "ioutil.ReadAll"
		return str2
	}
	md5 := fmt.Sprintf("%x", md5.Sum(body))

	return md5
}
