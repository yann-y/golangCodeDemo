package main

import (
	"fmt"
	shell "github.com/ipfs/go-ipfs-api"
	"os"
	"strings"
)

func main() {

	// Where your local node is running on localhost:5001
	sh := shell.NewShell("192.168.61.128:5001")
	cid, err := sh.Add(strings.NewReader("hello world!"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
	fmt.Printf("added %s", cid)

}
