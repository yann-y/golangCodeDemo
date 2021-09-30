package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

var host = "http://10.80.7.34"
var accessKey = "666888"
var token = "eyJhbGciOiJIUzI1NiJ9.eyJqdGkiOiIyNTciLCJpYXQiOjE2Mjk2OTkwODF9.54_WizUe9UEZxgCxMJdkTuZ96-Md4gSRz8AEQzEpobI"
var RegionID = "10027" //正式环境是30，测试环境是10027

const logName = "file_hash_updownload.log"

type GetUploadSignResp struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Data    struct {
		Sign       string `json:"sign"`
		Timestamp  int64  `json:"timestamp"`
		Expiretime int    `json:"expiretime"`
		NodeAddr   string `json:"nodeAddr"`
		Filepath   string `json:"filepath"`
		NodeIP     string `json:"nodeIp"`
		Uploadid   int    `json:"uploadid"`
		Userid     int    `json:"userid"`
		Regionid   string `json:"regionid"`
		Filename   string `json:"filename"`
		Filesize   int    `json:"filesize"`
	} `json:"data"`
	Total int `json:"total"`
}

type chunkJson struct {
	ChunkList []string `json:"chunkList"`
	Code      string   `json:"code"`
	State     int      `json:"state"`
}

var fileInfo struct {
	timestamp string //签名接口返回的时间戳
	hash      string //文件hash
	code      string //上传切片用的授权签名
	fileSize  int64  //文件大小
	chunkNums int    //文件切片的快数
}

func init() {
	writer, err := rotatelogs.New(
		logName+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(logName),
		rotatelogs.WithMaxAge(time.Hour*24),
		rotatelogs.WithRotationTime(time.Duration(604800)*time.Second),
	)
	if err != nil {
		panic(err)
	}
	multiWriter := io.MultiWriter(os.Stdout, writer)

	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(multiWriter)

	// Only log the DebugLevel severity or above.
	log.SetLevel(log.InfoLevel)

	// 显示行数
	log.SetReportCaller(true)
}

// 分片上传测试
func main() {
	localPath := ""
	mtywPath := ""
	gusp := getUploadSign(localPath, mtywPath)
	chunkList := getChunk(gusp, localPath)
	getFileChunk(chunkList.ChunkList, localPath)
	mergeChunk()
}

//localPath文件的本地路径,mtywPath云网的存储路径.获取上传的sign
func getUploadSign(localPath, mtywPath string) *GetUploadSignResp {
	fs := GetFileSize(localPath)
	fileInfo.fileSize = fs
	url := host + "/api/file/uploadIpfsSign?" + "balanceApply=true" + "&fileName=" + mtywPath + "&fileSize=" + strconv.FormatInt(fs, 10) +
		"&filepath=" + mtywPath + "&regionId=" + RegionID + "&flowId=" + "&accessKey=" + accessKey
	//log.Println(url)
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Set("token", token)
	if err != nil {
		log.Println("请求错误", err)
	}

	// 处理返回的结果
	response, _ := client.Do(request)
	// 获取body 内容
	body, _ := ioutil.ReadAll(response.Body)
	var gusp GetUploadSignResp
	err = json.Unmarshal(body, &gusp)
	if err != nil {
		log.Println("json gdsp 转换错误:", err, "body is: ", string(body))
		log.Error("json gdsp 转换错误", err, "body is: ", string(body))
		return &GetUploadSignResp{}
	}
	if len(gusp.Data.NodeIP) == 0 {
		log.Error("get uploadIpfsSign nodeip is nil: ", localPath)
		return &GetUploadSignResp{}
	}
	return &gusp
}

//上传切片
func getChunk(gusp *GetUploadSignResp, localPath string) *chunkJson {
	//获取文件hash
	fileMd5 := getFileMD5(localPath)
	fileInfo.hash = fileMd5
	url2 := gusp.Data.NodeAddr + "/fs/add?" + "expire=" + strconv.Itoa(gusp.Data.Expiretime) + "&sign=" + gusp.Data.Sign + "&filepath=" + gusp.Data.Filepath +
		"&nodeip=" + gusp.Data.NodeIP + "&timestamp=" + strconv.FormatInt(gusp.Data.Timestamp, 10) +
		"&regions=" + gusp.Data.Regionid + "&uploadid=" + strconv.Itoa(gusp.Data.Uploadid) + "&filesize=" + strconv.Itoa(gusp.Data.Filesize) +
		"&accesskey=" + accessKey + "&userid=" + strconv.Itoa(gusp.Data.Userid) + "&hash=" + fileMd5
	client := &http.Client{}
	request, err := http.NewRequest("GET", url2, nil)
	if err != nil {
		log.Println("请求错误", err)
	}
	// 处理返回的结果
	response, _ := client.Do(request)
	// 获取body 内容
	body, _ := ioutil.ReadAll(response.Body)
	var cJ chunkJson
	err = json.Unmarshal(body, &cJ)
	return &cJ
}

// GetFileSize 获取文件的大小
func GetFileSize(localPath string) int64 {
	fi, err := os.Stat(localPath)
	if err != nil {
		log.Println("stat file failed: ", err.Error())
		log.Error("stat file failed: ", err.Error())
		return 0
	}
	log.Info(localPath+" size is: ", fi.Size())

	return fi.Size()

}

//path文件的路径,获取文件的MD5
func getFileMD5(path string) (retMd5 string) {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(f)
	body, err := ioutil.ReadAll(f)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", md5.Sum(body))
}

//上传剩余的切片
func getFileChunk(chunkList []string, localPath string) {
	fileSize := fileInfo.fileSize
	chunkNums := int(math.Ceil(float64(fileSize / (1024 * 1024))))
	fileInfo.chunkNums = chunkNums

	mapList := make(map[int]struct{}, fileInfo.chunkNums)
	for _, v := range chunkList {
		atoi, _ := strconv.Atoi(v)
		mapList[atoi] = struct{}{}
	}
	for i := 0; i < chunkNums; i++ {
		if _, ok := mapList[i]; !ok {
			buf, err := ReadBlock(localPath, 1024*1024, i)
			if err != nil {
				continue
			}
			//上传buf
			url := "http://192.168.61.128:38888/fs/addChunk?" +
				"timestamp=" + fileInfo.timestamp +
				"&hash=" + fileInfo.hash +
				"&chunkName=" + strconv.Itoa(i) +
				"&code=" + fileInfo.code
			reader := bytes.NewReader(buf) //byte[]转io
			req, err := http.Post("POST", url, reader)
			if err != nil {
				log.Error(err)
				return
			}
			body, err := ioutil.ReadAll(req.Body)
			fmt.Printf(string(body))
		}

	}
}

// ReadBlock 读取相应位置的切片
func ReadBlock(filePth string, bufSize int, chunkName int) ([]byte, error) {
	f, err := os.OpenFile(filePth, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
		}
	}(f)
	buf := make([]byte, bufSize) //一次读取多少个字节
	_, err = f.Seek(int64(chunkName)*int64(bufSize), 0)
	_, err = f.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil

}

//合并分片
func mergeChunk() {
	url := "http://192.168.61.128:38888/fs/mergeChunk?" +
		"hash=" + fileInfo.hash +
		"&chunkNums=" + strconv.Itoa(fileInfo.chunkNums) +
		"&timestamp=" + fileInfo.timestamp +
		"&code=" + fileInfo.code
	fmt.Println(url)
	get, err := http.Get(url)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(get.Body)
	if err != nil {
		log.Error(err)
	}
	fmt.Println(string(body))
}
