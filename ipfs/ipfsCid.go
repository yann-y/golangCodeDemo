package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	cid "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"io/ioutil"
)

const (
	// File is the test file
	File = "/home/yanzq/Documents/新建文本文档.txt"
)

func main() {

	//p := cid.Prefix{
	//	Version: 0,
	//	Codec: cid.DagProtobuf,
	//	MhType: mh.SHA2_256,
	//	MhLength: -1, // default length
	//}
	buf, _ := hex.DecodeString("0487e270b71d439d5eaa3a8716a133dbb64d638e337b124dfdd81560e2543415") // hash
	mHashBuf, _ := mh.EncodeName(buf, "sha2-256")
	cidString := cid.NewCidV0(mHashBuf)

	fmt.Println(cidString)
	data, err := ioutil.ReadFile(File)
	sprintf := fmt.Sprintf("%x", sha256.Sum256(data))
	fmt.Println(sprintf)
	if err != nil {
		fmt.Println(err)
	}

	//
	buf, _ = hex.DecodeString("12200487e270b71d439d5eaa3a8716a133dbb64d638e337b124dfdd81560e2543415")
	cast, cid, err := cid.CidFromBytes(buf)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cast, cid)
	//fcid, err := p.Sum(data)
	//c, err := p.Sum([]byte("hello world"))
	//hash, err := mh.Sum(data, p.MhType, -1)
	//fmt.Println("hash",string(hash),"-",c)
	//if err != nil{
	//	fmt.Println(err)
	//}
	//
	//fmt.Println(fcid)
	//// Create a cid from a marshaled string

}
