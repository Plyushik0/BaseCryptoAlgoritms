package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование: go run main.go [server|client]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		startServer()
	case "client":
		mainClient()
	default:
		fmt.Println("Неверный аргумент. Используйте 'server' или 'client'")
		os.Exit(1)
	}
}