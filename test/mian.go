package main

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {

	log.Println("start...")
	/* 程序主体 */
	fmt.Printf(strMd5("./1.md"))

	log.Println("end...")

}

func strMd5(str string) (retMd5 string) {
	f, err := os.Open(str)
	if err != nil {
		return "md5 Open file err"
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println(err)
		}
	}(f)
	body, err := ioutil.ReadAll(f)
	if err != nil {
		return "ioutil.ReadAll"
	}
	sha1String := fmt.Sprintf("%x", sha1.Sum(body))

	fmt.Println(len(sha1String))
	fmt.Println(sha1String)
	return fmt.Sprintf("%x", md5.Sum(body))
}
