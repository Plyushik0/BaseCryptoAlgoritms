package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

var board = [9]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}

func main() {
	
	ln, _ := net.Listen("tcp", ":9000")
	fmt.Println("Сервер запущен на порту 9000")

	
	fmt.Println("Ждём первого игрока...")
	conn1, _ := ln.Accept()
	fmt.Fprintln(conn1, "Вы подключились! Ждём второго игрока...")

	
	fmt.Println("Ждём второго игрока...")
	conn2, _ := ln.Accept()
	fmt.Fprintln(conn2, "Вы подключились!")

	
	fmt.Fprintln(conn1, "Вы играете за X")
	fmt.Fprintln(conn2, "Вы играете за O")

	
	reader1 := bufio.NewReader(conn1)
	reader2 := bufio.NewReader(conn2)

	
	currentPlayer := 1

	for {
		
		sendBoard(conn1, conn2)

		var currentConn net.Conn
		var currentReader *bufio.Reader
		var currentSymbol string

		if currentPlayer == 1 {
			currentConn = conn1
			currentReader = reader1
			currentSymbol = "X"
			fmt.Fprintln(conn1, "Ваш ход! Введите номер клетки (1-9):")
			fmt.Fprintln(conn2, "Ход соперника, ждите...")
		} else {
			currentConn = conn2
			currentReader = reader2
			currentSymbol = "O"
			fmt.Fprintln(conn2, "Ваш ход! Введите номер клетки (1-9):")
			fmt.Fprintln(conn1, "Ход соперника, ждите...")
		}

		
		input, err := currentReader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(conn1, "Игрок отключился. Игра окончена.")
			fmt.Fprintln(conn2, "Игрок отключился. Игра окончена.")
			return
		}

		
		input = strings.TrimSpace(input)
		pos, err := strconv.Atoi(input)

		
		if err != nil || pos < 1 || pos > 9 {
			fmt.Fprintln(currentConn, "Некорректный ввод! Введите число от 1 до 9")
			continue
		}

		
		if board[pos-1] == "X" || board[pos-1] == "O" {
			fmt.Fprintln(currentConn, "Эта клетка уже занята!")
			continue
		}

		
		board[pos-1] = currentSymbol

		
		if checkWin() {
			sendBoard(conn1, conn2)
			fmt.Fprintln(currentConn, "Вы победили!")
			if currentPlayer == 1 {
				fmt.Fprintln(conn2, "Вы проиграли!")
			} else {
				fmt.Fprintln(conn1, "Вы проиграли!")
			}
			return
		}

		
		if checkDraw() {
			sendBoard(conn1, conn2)
			fmt.Fprintln(conn1, "Ничья!")
			fmt.Fprintln(conn2, "Ничья!")
			return
		}

		
		if currentPlayer == 1 {
			currentPlayer = 2
		} else {
			currentPlayer = 1
		}
	}
}

func sendBoard(conn1 net.Conn, conn2 net.Conn) {
	b := fmt.Sprintf("\n %s | %s | %s \n---|---|---\n %s | %s | %s \n---|---|---\n %s | %s | %s \n",
		board[0], board[1], board[2],
		board[3], board[4], board[5],
		board[6], board[7], board[8])
	fmt.Fprintln(conn1, b)
	fmt.Fprintln(conn2, b)
}

func checkWin() bool {
	
	wins := [8][3]int{
		{0, 1, 2}, 
		{3, 4, 5}, 
		{6, 7, 8}, 
		{0, 3, 6}, 
		{1, 4, 7}, 
		{2, 5, 8}, 
		{0, 4, 8}, 
		{2, 4, 6}, 
	}

	for _, w := range wins {
		if board[w[0]] == board[w[1]] && board[w[1]] == board[w[2]] {
			return true
		}
	}
	return false
}

func checkDraw() bool {
	for _, cell := range board {
		if cell != "X" && cell != "O" {
			return false
		}
	}
	return true
}
