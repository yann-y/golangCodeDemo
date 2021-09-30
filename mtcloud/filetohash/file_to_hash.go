package main

import (
	"bufio"
	md52 "crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
)

//读取某个文件夹下的所有文件，将每个文件里的内容用md5加密后写入txt文件。
func main() {
	path := "/test2/test3/"
	uploadFile := path + "upload.txt"
	ReadFiletoHash(path, uploadFile)

	newpath := "/new/"
	downloadFile := newpath + "download.txt"
	ReadFiletoHash(newpath, downloadFile)
}

func ReadFiletoHash(path string, txtFile string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println("ReadDir failed: ", err)
		return
	}
	for _, file := range files {
		data, err := ioutil.ReadFile(path + file.Name())
		if err != nil {
			fmt.Println("ReadFile failed: ", err)
		}
		md5 := fmt.Sprintf("%x", md52.Sum(data))
		WriteFile(path+file.Name()+"---"+md5, txtFile)
	}
}

func WriteFile(content string, path string) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("文件打开失败：", err)
	}
	defer file.Close()
	//写入文件时，使用带缓存的 *Writer
	write := bufio.NewWriter(file)
	write.WriteString(content + " \n")
	//Flush将缓存的文件真正写入到文件中
	write.Flush()
}
