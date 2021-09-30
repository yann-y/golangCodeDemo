package main

import (
	"bufio"
	"bytes"
	md52 "crypto/md5"
	"encoding/json"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var oldHashMap map[string]string //上传文件hash记录
var newHashMap map[string]string //下载文件hash记录
var host = "https://pro.mty.wang"

//"http://10.80.7.34"，测试环境地址,需要挂VPN
//"https://pro.mty.wang",正式环境地址，不需要VPN

//var uploadHost = "https://wznode3.dss.mty.wang:28089"
var accessKey = "666888"

//正式环境的accessKey = "666888"
var token = "eyJhbGciOiJIUzI1NiJ9.eyJqdGkiOiIyNTciLCJpYXQiOjE2Mjk2OTkwODF9.54_WizUe9UEZxgCxMJdkTuZ96-Md4gSRz8AEQzEpobI"

//正式环境测试用的token = "eyJhbGciOiJIUzI1NiJ9.eyJqdGkiOiIyNTciLCJpYXQiOjE2Mjk2OTkwODF9.54_WizUe9UEZxgCxMJdkTuZ96-Md4gSRz8AEQzEpobI"
var UploadFiles []string
var RegionID = "30"      //正式环境是30，测试环境是10027
var IsConcurrency = true //是否并发，false否、true并发

const logName = "file_hash_updownload.log"

var append_upload_lock sync.Locker

type GetDownloadSignResp struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Data    struct {
		Sign      string `json:"sign"`
		Nodeip    string `json:"nodeip"`
		NodeAddr  string `json:"nodeAddr"`
		Filename  string `json:"filename"`
		Cid       string `json:"cid"`
		Timestamp int64  `json:"timestamp"`
		UID       int    `json:"uid"`
	} `json:"data"`
	Total int `json:"total"`
}

type GetFscatResp struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Data    string `json:"data"`
}

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

type PostFsaddResp struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Data    string `json:"data"`
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

func main() {
	start := time.Now()
	args := os.Args
	ConcurrencyCount, _ := strconv.Atoi(args[2])
	var UploadConcurrencyCount = ConcurrencyCount   //上传文件并发个数
	var DownloadConcurrencyCount = ConcurrencyCount //下载文件并发个数
	oldHashMap = make(map[string]string)
	newHashMap = make(map[string]string)
	//1,读取文件，生成hash值
	path := args[1] // "/test2/test3/"
	mtywPath := "/test2/test3/"
	uploadFile := path + "upload.txt"
	ReadFiletoHash(path, uploadFile)

	//2，上传文件
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println("ReadDir failed: ", err)
		return
	}

	if IsConcurrency {
		//并发版本
		wg := sync.WaitGroup{}
		partCount := len(files) / UploadConcurrencyCount

		for i := 0; i < UploadConcurrencyCount; i++ {
			//part := files[0:300]
			var part []fs.FileInfo

			if i == 0 {
				//第一位
				part = files[i:partCount]
				log.Info("upload i = ", i, ",upload part = ", part)
			} else if i == UploadConcurrencyCount-1 {
				//最后一位
				part = files[i*partCount:]
				log.Info("upload i = ", i, ",upload part = ", part)
			} else {
				//中间
				part = files[i*partCount : i*partCount+partCount] //panic: runtime error: slice bounds out of range [:70] with capacity 64
				log.Info("upload i = ", i, ",upload part = ", part)
			}

			wg.Add(1)
			go func(part []fs.FileInfo) {
				for i := 0; i < len(part); i++ {
					HttpUploadFile(path+part[i].Name(), mtywPath+part[i].Name())
				}
				wg.Done()
			}(part)
		}

		wg.Wait()
	} else {
		for _, file := range files {
			HttpUploadFile(path+file.Name(), mtywPath+file.Name())
		}
	}

	//3，下载文件
	newpath := "/new/"
	if IsConcurrency {
		//并发版本
		wg := sync.WaitGroup{}
		partCount := len(UploadFiles) / DownloadConcurrencyCount
		for i := 0; i < DownloadConcurrencyCount; i++ {
			//part := files[0:300]
			var part []string

			if i == 0 {
				part = UploadFiles[0:partCount]
				log.Info("download i = ", i, ",download part = ", part)
			} else if i == DownloadConcurrencyCount-1 {
				part = UploadFiles[i*partCount:]
				log.Info("download i = ", i, ",download part = ", part)
			} else {
				part = UploadFiles[i*partCount : i*partCount+partCount]
				log.Info("download i = ", i, ",download part = ", part)
			}

			wg.Add(1)
			go func(part []string) {
				for i := 0; i < len(part); i++ {
					HttpDownloadFile(part[i], newpath)
				}
				wg.Done()
			}(part)
		}

		wg.Wait()

		//results := make(chan string, len(files)) //容量为files数量的名为results通道
		//for _,file := range UploadFiles {
		//	for i := 1;i <= DownloadConcurrencyCount;i++ {
		//		go workerDownload(i,results,file,newpath)
		//	}
		//}

		//for a := 1; a <= len(files); a++ {
		//	<-results //接收元素值并将它丢弃
		//}
	} else {
		for _, file := range UploadFiles {
			HttpDownloadFile(file, newpath)
		}
	}

	//4,读取下载下来的文件，生成hash
	downloadFile := newpath + "download.txt"
	ReadFiletoHash(newpath, downloadFile)

	//5,计算成功和失败个数
	cf, df := IsEqualHash()
	log.Printf("不一致个数：%d、下载失败个数：%d", cf, df)

	cost := time.Since(start)
	log.Printf("cost=[%s]", cost)
}

func workerUpload(id int, results chan string, file string) {
	results <- "Do Success"
	HttpUploadFile(file, file)

	log.Info("upload worker: ", id, " started file:", file)
	time.Sleep(time.Second)
	log.Info("upload worker: ", id, " finished file:", file)
	//<- results
}

func workerDownload(id int, results chan string, file string, base string) {
	results <- "Do Success"
	HttpDownloadFile(file, base)

	log.Info("download worker: ", id, " started file:", file)
	time.Sleep(time.Second)
	log.Info("download worker: ", id, " finished file:", file)
	//<- results
}

//func ReadFiletoHash(path string) {
//	oldHashMap = make(map[string]string)
//	files,err := ioutil.ReadDir(path)
//	if err != nil {
//		log.Println("ReadDir failed: ",err)
//		return
//	}
//	for _,file := range files {
//		data,err := ioutil.ReadFile(path+file.Name())
//		if err != nil {
//			log.Println("ReadFile failed: ",err)
//		}
//		md5 := log.Sprintf("%x",md52.Sum(data))
//		oldHashMap[file.Name()] = md5
//	}
//}
//
//func NewReadFiletoHash(path string) {
//	files,err := ioutil.ReadDir(path)
//	if err != nil {
//		log.Println("ReadDir failed: ",err)
//		return
//	}
//	for _,file := range files {
//		data,err := ioutil.ReadFile(path+file.Name())
//		if err != nil {
//			log.Println("ReadFile failed: ",err)
//		}
//		md5 := log.Sprintf("%x",md52.Sum(data))
//		newHashMap[file.Name()] = md5
//	}
//}

//path本地存储路径，mtywPath云网上的存储路径
func HttpUploadFile(path string, mtywPath string) {
	//0,计算文件大小，单位字节
	fs := GetFileSize(path)
	//3、获取上传授权签名
	url := host + "/api/file/uploadIpfsSign?" + "balanceApply=true" + "&fileName=" + mtywPath + "&fileSize=" + strconv.FormatInt(fs, 10) +
		"&filepath=" + mtywPath + "&regionId=" + RegionID + "&flowId=" + "&accessKey=" + accessKey
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", url, nil)
	reqest.Header.Set("token", token)
	if err != nil {
		log.Println("请求错误", err)
	}

	// 处理返回的结果
	response, _ := client.Do(reqest)
	// 获取body 内容
	body, _ := ioutil.ReadAll(response.Body)
	var gusp GetUploadSignResp
	err = json.Unmarshal(body, &gusp)
	if err != nil {
		log.Println("json gdsp 转换错误:", err, "body is: ", string(body))
		log.Error("json gdsp 转换错误", err, "body is: ", string(body))
		return
	}
	if len(gusp.Data.NodeIP) == 0 {
		log.Error("get uploadIpfsSign nodeip is nil: ", path)
		return
	}

	//4、上传文件
	url2 := gusp.Data.NodeAddr + "/fs/add?" + "expire=" + strconv.Itoa(gusp.Data.Expiretime) + "&sign=" + gusp.Data.Sign + "&filepath=" + gusp.Data.Filepath +
		"&nodeip=" + gusp.Data.NodeIP + "&timestamp=" + strconv.FormatInt(gusp.Data.Timestamp, 10) +
		"&regions=" + gusp.Data.Regionid + "&uploadid=" + strconv.Itoa(gusp.Data.Uploadid) + "&filesize=" + strconv.Itoa(gusp.Data.Filesize) +
		"&accesskey=" + accessKey + "&userid=" + strconv.Itoa(gusp.Data.Userid)

	// 处理返回的结果
	//response2, err := client2.Do(reqest2)
	response2, err := postFile(path, url2)
	if err != nil {
		log.Println("response2 failed:", err)
	}
	defer response2.Body.Close()
	// 获取body 内容
	body2, _ := ioutil.ReadAll(response2.Body)
	var pfr PostFsaddResp
	err = json.Unmarshal(body2, &pfr)
	if err != nil {
		log.Println("json gdsp 转换错误", err)
	}

	if !pfr.Success {
		log.Println("文件上传失败:", gusp.Data.Filepath)
	}

	//append_upload_lock.Lock()
	UploadFiles = append(UploadFiles, gusp.Data.Filepath)
	//append_upload_lock.Unlock()
}

func HttpDownloadFile(path string, base string) {
	//1、获取下载授权签名
	url := host + "/api/file/ipfsDownloadSign?" + "filepath=" + path + "&accessKey=" + accessKey
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", url, nil)
	reqest.Header.Set("token", token)
	if err != nil {
		log.Println("ipfsDownloadSign 请求错误: ", err)
	}

	// 处理返回的结果
	response, _ := client.Do(reqest)
	// 获取body 内容
	body, _ := ioutil.ReadAll(response.Body)

	var gdsp GetDownloadSignResp

	err = json.Unmarshal(body, &gdsp)
	if err != nil {
		log.Println("json gdsp 转换错误", err, "body is: ", string(body))
		log.Error("json gdsp 转换错误", err, "body is: ", string(body))
		return
	}
	if len(gdsp.Data.Nodeip) == 0 {
		log.Error("get ipfsDownloadSign nodeip is nil: ", path)
		return
	}

	//2、下载文件
	url2 := "https://" + gdsp.Data.Nodeip + ":28089" + "/fs/cat?" + "sign=" + gdsp.Data.Sign + "&timestamp=" + strconv.FormatInt(gdsp.Data.Timestamp, 10) +
		"&nodeip=" + gdsp.Data.Nodeip + "&cid=" + gdsp.Data.Cid + "&userid=" + strconv.Itoa(gdsp.Data.UID) +
		"&filename=" + gdsp.Data.Filename
	client2 := &http.Client{}
	reqest2, err := http.NewRequest("GET", url2, nil)
	reqest2.Header.Set("token", token)
	if err != nil {
		log.Println("请求错误", err)
	}

	// 处理返回的结果
	response2, err := client2.Do(reqest2)
	if err != nil {
		log.Println("DO reqest2 failed: ", err)
	}
	defer response2.Body.Close()
	// 获取body 内容
	//body2, err := ioutil.ReadAll(response2.Body)  //读取一次就清空？
	//if err != nil {
	//	log.Println("ReadAll response2.Body is: ",err)
	//}
	//var gfp GetFscatResp
	//err = json.Unmarshal(body2, &gfp)
	//if err != nil {
	//	log.Println("json gdsp 转换错误", err)
	//}
	//log.Println("body2 is: ",string(body2))

	//创建文件
	out, err := os.Create(base + gdsp.Data.Filename)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	_, ioerr := io.Copy(out, response2.Body)
	if ioerr != nil {
		panic(err)
	}
}

func reqest(method, url string) map[string]interface{} {
	// 生成client参数为默认
	client := &http.Client{}

	// 提交请求
	reqest, err := http.NewRequest(method, url, nil)
	reqest.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	reqest.Header.Set("Accept-Charset", "GBK,utf-8;q=0.7,*;q=0.3")
	reqest.Header.Set("Accept-Encoding", "gzip,deflate,sdch")
	reqest.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	reqest.Header.Set("Cache-Control", "max-age=0")
	reqest.Header.Set("Connection", "keep-alive")
	//reqest.Header.Set("Content-Type", "multipart/form-data")
	reqest.Header.Set("Content-Type", "text/html")
	// 错误处理
	if err != nil {
		log.Println("请求错误", err)
	}
	// 处理返回的结果
	response, _ := client.Do(reqest)
	// 获取body 内容
	body, _ := ioutil.ReadAll(response.Body)
	// 存储返回结果
	result := make(map[string]interface{}, 0)
	// 把json转换为map
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Println("json 转换错误", err)
	}

	//返回的状态码
	status := response.StatusCode

	if status != 200 {
		log.Println("请求错误地址  错误编码：", status)
	}

	//defer reqest.Body.Close()

	return result
}

func IsEqualHash() (int, int) {
	countFailed := 0
	downloadFailed := 0
	if len(oldHashMap) != len(newHashMap) {
		downloadFailed = len(oldHashMap) - len(newHashMap)
	}
	for name, oldHash := range oldHashMap {
		if oldHash != newHashMap[name] {
			countFailed++
		}
	}

	return countFailed, downloadFailed
}

func GetFileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		log.Println("stat file failed: ", err.Error())
		log.Error("stat file failed: ", err.Error())
		return 0
	}
	log.Info(path+" size is: ", fi.Size())

	return fi.Size()
}

func MapDataToFile(mapData map[string]string, path string) {
	if mapData == nil {
		log.Println("map为空，返回")
		return
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("文件打开失败", err)
	}
	//及时关闭file句柄
	defer file.Close()
	write := bufio.NewWriter(file)
	for fileName, md5 := range mapData {
		//err := ioutil.WriteFile(path,[]byte(fileName + "---" + md5 + "\n"),0666)
		//if err != nil {
		//	log.Println("MapDataToFile failed:",err)
		//}
		write.WriteString(fileName + "---" + md5 + "\n")
	}
}

func postFile(filename string, target_url string) (*http.Response, error) {
	body_buf := bytes.NewBufferString("")
	body_writer := multipart.NewWriter(body_buf)

	// use the body_writer to write the Part headers to the buffer
	_, err := body_writer.CreateFormFile("userfile", filename)
	if err != nil {
		log.Println("error writing to buffer")
		return nil, err
	}

	// the file data will be the second part of the body
	fh, err := os.Open(filename)
	if err != nil {
		log.Println("error opening file")
		return nil, err
	}
	// need to know the boundary to properly close the part myself.
	boundary := body_writer.Boundary()
	//close_string := log.Sprintf("\r\n--%s--\r\n", boundary)
	close_buf := bytes.NewBufferString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

	// use multi-reader to defer the reading of the file data until
	// writing to the socket buffer.
	request_reader := io.MultiReader(body_buf, fh, close_buf)
	fi, err := fh.Stat()
	if err != nil {
		log.Printf("Error Stating file: %s", filename)
		return nil, err
	}
	req, err := http.NewRequest("POST", target_url, request_reader)
	if err != nil {
		return nil, err
	}

	//TODO:需要判断文件类型，填入对应的Content-Type
	// Set headers for multipart, and Content Length
	req.Header.Add("Content-Type", "multipart/form-data; boundary="+boundary)
	req.ContentLength = fi.Size() + int64(body_buf.Len()) + int64(close_buf.Len())

	return http.DefaultClient.Do(req)
}

func ReadFiletoHash(path string, txtFile string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println("ReadDir failed: ", err)
		return
	}
	for _, file := range files {
		data, err := ioutil.ReadFile(path + file.Name())
		if err != nil {
			log.Println("ReadFile failed: ", err)
		}
		md5 := fmt.Sprintf("%x", md52.Sum(data))
		WriteFile(path+file.Name()+"---"+md5, txtFile)

		if strings.Contains(txtFile, "upload.txt") {
			oldHashMap[file.Name()] = md5
		} else if strings.Contains(txtFile, "download.txt") {
			newHashMap[file.Name()] = md5
		}

	}
}

func WriteFile(content string, path string) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Println("文件打开失败：", err)
	}
	defer file.Close()
	//写入文件时，使用带缓存的 *Writer
	write := bufio.NewWriter(file)
	write.WriteString(content + " \n")
	//Flush将缓存的文件真正写入到文件中
	write.Flush()
}
