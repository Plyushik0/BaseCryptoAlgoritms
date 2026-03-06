package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		fmt.Println("Не удалось подключиться к серверу!")
		return
	}
	defer conn.Close()

	// Запускаем отдельную горутину которая читает сообщения от сервера
	// и печатает их на экран
	go func() {
		scanner := bufio.NewScanner(conn)
		// scanner.Scan() — блокируется и ждёт пока пользователь не нажмёт Enter. Scanner добавляет возможность читать построчно.
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		fmt.Println("Соединение с сервером закрыто")
		os.Exit(0)
	}()

	// ввод пользователя
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Fprintln(conn, scanner.Text())
	}
}
