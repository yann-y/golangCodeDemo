package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"sync"
)

type mpStruct struct {
	ChunkList []string `json:"chunkList"`
	State     int      `json:"state"`
	FileName  string   `json:"fileName"`
	FileSize  int64    `json:"fileSize"`
}

var rwLock sync.RWMutex

func createFile(fileSize int64, filePath string) {
	f, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := f.Truncate(fileSize); err != nil {
		log.Fatal(err)
	}
}
func main() {
	router := gin.Default()
	router.Use(Cors)
	// 获取缓存信息
	router.GET("/checkChunk", func(c *gin.Context) {
		hash := c.Query("hash")
		fileName := c.Query("filename")
		fileSize := c.Query("fileSize")
		hashPath := fmt.Sprintf("./uploadFile/%s", hash) // 缓存文件夹
		mpJsonPath := fmt.Sprintf("%s/%s", hashPath, "mp.json")
		isExistJson, err := PathExists(mpJsonPath)
		if err != nil {
			fmt.Println("获取hash路径错误", err)
		}
		if isExistJson {
			// 读取json
			bytes, err := ioutil.ReadFile(mpJsonPath)
			if err != nil {
				fmt.Println("读取json文件失败", err)
				return
			}
			mp := &mpStruct{}
			err = json.Unmarshal(bytes, mp)
			if err != nil {
				fmt.Println("解析数据失败", err)
				return
			}
			//fmt.Printf("%+v\n", mp)
			c.JSON(200, gin.H{
				"state":     mp.State,
				"chunkList": mp.ChunkList,
			})
			// 保存文件信息
			mp.FileName = fileName
			mp.FileSize, _ = strconv.ParseInt(fileSize, 10, 64)
			bytes, err = json.Marshal(mp)
			err = ioutil.WriteFile(mpJsonPath, bytes, os.ModePerm)
			log.Println(err)
		} else {

			c.JSON(200, gin.H{
				"state":     0,
				"chunkList": []string{},
			})
			mp := &mpStruct{}
			mp.State = 0
			mp.ChunkList = []string{}
			mp.FileSize, _ = strconv.ParseInt(fileSize, 10, 64)
			mp.FileName = fileName
			_, err := os.Create(fileName)
			if err != nil {
				fmt.Println("创建文件失败")
			}
			bytes, err := json.Marshal(mp)
			err = ioutil.WriteFile(mpJsonPath, bytes, os.ModePerm)
			fmt.Printf("%+v\n", mp)
		}

	})

	router.POST("/uploadChunk", func(c *gin.Context) {
		fileHash := c.PostForm("hash")
		file, err := c.FormFile("file")
		//body := c.Request.Body
		hashPath := fmt.Sprintf("./uploadFile/%s", fileHash)
		if err != nil {
			fmt.Println("获取上传文件失败", err)
		}

		isExistPath, err := PathExists(hashPath)
		if err != nil {
			fmt.Println("获取hash路径错误", err)
		}
		mpJsonPath := fmt.Sprintf("%s/%s", hashPath, "mp.json")
		if !isExistPath {
			err := os.Mkdir(hashPath, os.ModePerm)
			if err != nil {
				fmt.Println("创建目录失败")
			}
			_, err = os.Create(mpJsonPath)
			if err != nil {
				fmt.Println("创建mpjson失败")
			}
		}

		err = c.SaveUploadedFile(file, fmt.Sprintf("./uploadFile/%s/%s", fileHash, file.Filename))
		isExistJson, err := PathExists(mpJsonPath)
		if isExistJson {
			//读取json
			mp, err := rwJson(mpJsonPath, file.Filename)
			if err != nil {
				fmt.Println("读取json失败！")
			}
			// start 不保存切片，直接偏移写入
			//1.先判断文件是否存在，不存在创建文件
			//filePath := fmt.Sprintf("./uploadFile/%s/%s", fileHash, mp.FileName)
			//isExistFile, err := PathExists(filePath)
			//if !isExistFile {
			// 文件不存在创建空白文件
			//	createFile(mp.FileSize, filePath)
			//}
			//fileLoad, _ := os.OpenFile(filePath, os.O_RDWR, 0)
			//defer fileLoad.Close()
			//i, _ := strconv.ParseInt(file.Filename, 10, 64)
			//_, err = fileLoad.Seek(i*2*1024*1024, 0)
			//fileLoad.Write(file.Open())

			// end
			if err != nil {
				fmt.Println("写入json错误")
			}
			c.JSON(200, gin.H{
				"chunkList": mp.ChunkList,
			})
		}

	})

	router.GET("megerChunk", func(c *gin.Context) {
		hash := c.Query("hash")
		fileName := c.Query("fileName")
		hashPath := fmt.Sprintf("./uploadFile/%s", hash)
		isExistPath, err := PathExists(hashPath)
		if err != nil {
			fmt.Println("获取hash路径错误", err)
		}
		if !isExistPath {
			c.JSON(400, gin.H{
				"message": "文件夹不存在",
			})
			return
		}
		isExistFile, err := PathExists(hashPath + "/" + fileName)
		if err != nil {
			fmt.Println("获取hash路径文件错误", err)
		}
		fmt.Println("文件是否存在", isExistFile)
		if isExistFile {
			c.JSON(200, gin.H{
				"fileUrl": fmt.Sprintf("http://127.0.0.1:9999/uploadFile/%s/%s", hash, fileName),
			})
			return
		}
		mpJsonPath := hashPath + "/" + "mp.json"
		mp := &mpStruct{}
		bytes, err := ioutil.ReadFile(mpJsonPath)
		err = json.Unmarshal(bytes, mp)

		//files, err := ioutil.ReadDir(hashPath)
		files := len(mp.ChunkList)
		if err != nil {
			fmt.Println("合并文件读取失败", err)
		}
		complateFile, err := os.Create(hashPath + "/" + fileName)
		defer complateFile.Close()
		for i := 0; i < files; i++ {
			fileBuffer, err := ioutil.ReadFile(hashPath + "/" + strconv.Itoa(i))
			if err != nil {
				fmt.Println("文件打开错误", err)
			}
			complateFile.Write(fileBuffer)
		}
		newMD5 := strMd5(hashPath + "/" + fileName)
		if hash == newMD5 {
			fmt.Println("md5一致")
		} else {
			fmt.Println("md5不一致", newMD5, "!=", hash)
		}
		mp.State = 1
		bytes, err = json.Marshal(mp)
		err = ioutil.WriteFile(mpJsonPath, bytes, os.ModePerm)
		c.JSON(200, gin.H{
			"fileUrl": fmt.Sprintf("http://127.0.0.1:9999/uploadFile/%s/%s", hash, fileName),
		})

	})

	err := router.Run("127.0.0.1:9999")
	if err != nil {
		fmt.Println("服务启动失败")
		return
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func strMd5(str string) (retMd5 string) {
	f, err := os.Open(str)
	if err != nil {
		return "md5 Open file err"
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(f)
	body, err := ioutil.ReadAll(f)
	if err != nil {
		return "ioutil.ReadAll"
	}
	return fmt.Sprintf("%x", md5.Sum(body))
}
func Cors(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	c.Next()
}
func rwJson(path, fileName string) (*mpStruct, error) {
	rwLock.Lock()         //获取写锁
	defer rwLock.Unlock() //释放写锁
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("%s读取错误", path)
		return nil, err
	}
	mp := &mpStruct{}
	err = json.Unmarshal(bytes, mp)
	oldLen := len(mp.ChunkList)
	addTrue := true // 去重，改片段存在就不添加
	for _, value := range mp.ChunkList {
		if value == fileName {
			addTrue = false
		}
	}
	if addTrue {
		mp.ChunkList = append(mp.ChunkList, fileName)
	}
	if len(mp.ChunkList) == oldLen {
		fmt.Println(fileName)
		fmt.Printf("%+v \n", mp.ChunkList)
	}
	bytes, err = json.Marshal(mp)
	err = ioutil.WriteFile(path, bytes, os.ModePerm)
	return mp, err
}
