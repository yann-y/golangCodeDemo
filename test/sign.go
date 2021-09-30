package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

const secret = "KaSToOM10IsNooLfhnc8aSgD7TOPPgZcmMt2c/8zvwIDAQAB"
const accessId = "666888"

func main() {
	//// mtc  sign生成方式
	//str := "accesskey=666888timestamp=1632734195042method=/fs/addexpire=1821110399filepath=/test1/test1.jpg" +
	//	"filesize=47430nodeip=test27.dss.mty.wangregions=10027uploadid=19444userid=261" +
	//	"KaSToOM10IsNooLfhnc8aSgD7TOPPgZcmMt2c/8zvwIDAQAB"
	//sign := sha256.Sum256([]byte(str))
	//goSignString := hex.EncodeToString(sign[:])
	//fmt.Printf(goSignString)
	hash := getCode("pwd", 12)
	fmt.Println(hash, "  ", len(hash), hash[64:])
}
func getCode(dir string, fileSize int64) string {
	//时间戳+上传id+
	newNow := time.Now()
	newNow = newNow.Add(time.Hour * 2)
	tstring := fmt.Sprintf("%d", newNow.Unix())
	fmt.Println(tstring)
	enc := fmt.Sprintf("accesskey=%sdir=%sfileSize=%dsecret=%s", accessId, dir, fileSize, secret)
	sign := sha256.Sum256([]byte(enc))
	goSignString := hex.EncodeToString(sign[:]) + tstring
	return goSignString
}
