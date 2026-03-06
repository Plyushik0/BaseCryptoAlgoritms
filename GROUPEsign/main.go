package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование: go run main.go [params|leader|member|ts|verify]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "params":
		if err := generateGroupParameters(2); err != nil {
			fmt.Printf("Ошибка генерации параметров: %v\n", err)
		}
	case "leader":
		if err := startLeaderServer(); err != nil {
			fmt.Printf("Ошибка сервера лидера: %v\n", err)
		}
	case "member":
		if len(os.Args) < 3 {
			fmt.Println("Использование: go run main.go member <member_id>")
			return
		}
		memberID, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Printf("Ошибка преобразования member_id: %v\n", err)
			return
		}
		if err := memberSignDocument(memberID, "document_to_sign.txt", 2); err != nil {
			fmt.Printf("Ошибка подписи участника: %v\n", err)
		}
	case "ts":
		if err := startTimestampServer(); err != nil {
			fmt.Printf("Ошибка сервера TS: %v\n", err)
		}
	case "verify":
		if err := verifyGroupSignature("document_to_sign.txt", "group_signature.cades"); err != nil {
			fmt.Printf("Ошибка проверки подписи: %v\n", err)
		}
	default:
		fmt.Println("Неверный аргумент")
	}
}