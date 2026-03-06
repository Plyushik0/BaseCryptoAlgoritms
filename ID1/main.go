package main

import (
    "bufio"

    "fmt"
    "os"
    "strings"
    "time"
    
)



func main() {
    fmt.Println("Протокол")

    if _, err := os.Stat(KEY_FILE); os.IsNotExist(err) {
        if err := GenerateAndSaveKey(); err != nil {
            os.Exit(1)
        }
    }

    fmt.Println("\nВведите сообщения для стороны A:")
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("Введите сообщение M1 (отправляется A, внутри шифрования): ")
    m1, _ := reader.ReadString('\n')
    m1 = strings.TrimSpace(m1)

    fmt.Print("Введите сообщение M2 (отправляется A, вместе с шифрованием): ")
    m2, _ := reader.ReadString('\n')
    m2 = strings.TrimSpace(m2)

    var useTimestamp bool
    for {
        fmt.Print("Выберите тип идентификатора (1 для Timestamp, 2 для Random number): ")
        choice, _ := reader.ReadString('\n')
        choice = strings.TrimSpace(choice)
        if choice == "1" {
            useTimestamp = true
            fmt.Println("Используется Timestamp в качестве идентификатора.")
            break
        } else if choice == "2" {
            useTimestamp = false
            fmt.Println("Используется случайное число в качестве идентификатора.")
            break
        } else {
            fmt.Println("Неверный выбор. Введите 1 или 2.")
        }
    }

    fmt.Println("\nВведите сообщения для стороны B:")
    fmt.Print("Введите сообщение M3 (отправляется B, внутри шифрования): ")
    m3, _ := reader.ReadString('\n')
    m3 = strings.TrimSpace(m3)

    fmt.Print("Введите сообщение M4 (отправляется B, вместе с шифрованием): ")
    m4, _ := reader.ReadString('\n')
    m4 = strings.TrimSpace(m4)

    go func() {
        if !RunUserB(m3, m4) {
            println("Сторона B завершилась с ошибкой.")
            os.Exit(1)
        }
    }()

    time.Sleep(1 * time.Second)

    if !RunUserA(m1, m2, useTimestamp) {
        println("Сторона A завершилась с ошибкой.")
        os.Exit(1)
    }
}