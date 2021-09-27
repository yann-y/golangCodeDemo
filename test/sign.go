package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type AddParams struct {
	UserId   int64   `param:"userid"`
	NodeIp   string  `param:"nodeip"`
	FilePath string  `param:"filepath"`
	FileSize int64   `param:"filesize"`
	Regions  []int64 `param:"regions"`
	UploadId int64   `param:"uploadid"`
	Expire   int64   `param:"expire"`
}

func main() {
	//
	accessKey := "666888"
	tstring := "1632382090278"
	method := "get"
	const secret = "KaSToOM10IsNooLfhnc8aSgD7TOPPgZcmMt2c/8zvwIDAQAB"
	paramsString := "/fs/add?expire=1821110399&sign=373d90afcb6502885cb297be6847bdcfa1d55ffcca6bd5cc1a63b4d6081f9e31&filepath=/test2/78.png&userid=261&nodeip=test27.dss.mty.wang&timestamp=1632382090278&regions=10027&uploadid=18240&filesize=34470&accesskey=666888&flowId="
	//numFiled := out.NumField()
	//pl := make(paramsSlice, numFiled)
	//sort.Sort(slice)
	//paramsString := strings.Join(slice.Slice(), "")
	enc := fmt.Sprintf("accesskey=%stimestamp=%smethod=%s%s%s", accessKey, tstring, method, paramsString, secret)
	sign := sha256.Sum256([]byte(enc))
	goSignString := hex.EncodeToString(sign[:])
	fmt.Printf(goSignString)
}
