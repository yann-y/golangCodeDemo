package main

import (
	"fmt"
	"os"
	"path"
)

func main() {

	file, _ := os.Getwd()
	file = path.Join(file, "mp.json")
	fmt.Printf(file)
}
