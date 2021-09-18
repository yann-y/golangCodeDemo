package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

type mpStruct struct {
	ChunkList []string `json:"chunkList"`
	State     int      `json:"state"`
}

func createFile(size int64, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := f.Truncate(size); err != nil {
		log.Fatal(err)
	}
	// Output:
	//
	finfo, _ := os.Stat("./" + filename)
	log.Println(finfo.Size())

	fi, _ := f.Stat()
	log.Println("f Stat:", fi.Size())
}
func main() {
	//createFile(11619280, "1.txt")
	hash := "9ac1d5fa25d6b10ae3ba4168e90ab7f3"

	file, _ := os.OpenFile("./1.txt", os.O_RDWR, 0)
	defer file.Close()

	//fmt.Println(seek)
	//input, err := ioutil.ReadFile("./uploadFile/" + hash + "/0")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	mpJsonPath := "./uploadFile/" + hash + "/" + "mp.json"
	mp := &mpStruct{}
	bytes, err := ioutil.ReadFile(mpJsonPath)
	err = json.Unmarshal(bytes, mp)
	if err != nil {
		return
	}

	for i := range mp.ChunkList {
		fmt.Println(i)
		_, err := file.Seek(int64(i*2*1024*1024), 0)
		if err != nil {
			return
		}
		input, err := ioutil.ReadFile("./uploadFile/" + hash + "/" + strconv.Itoa(i))
		if err != nil {
			fmt.Println(err)
			return
		}
		_, err = file.Write(input)
		if err != nil {
			return
		}
	}

}
