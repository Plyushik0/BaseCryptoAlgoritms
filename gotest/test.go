package main

import (
	"fmt"
	"os"
)

func main() {
	for i, args := range os.Args[1:] {
		fmt.Println(i, args)
	}
}
