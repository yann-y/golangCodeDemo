package main

import (
	"fmt"
	shell "github.com/ipfs/go-ipfs-api"
	"os"
)

func main() {

	// Where your local node is running on localhost:5001
	sh := shell.NewShell("localhost:5001")
	f, err := os.OpenFile("./README.md", os.O_RDONLY, 0600)
	cid, err := sh.Add(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
	fmt.Printf("added %s", cid)

}
