package main

import (
    "bufio"

    "fmt"
    "os"
    "strings"
    "time"

)

func main() {
    fmt.Println("--- Симуляция трёхпроходного протокола идентификации ---")

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
    if m1 == "" {
        fmt.Println("Ошибка: M1 не может быть пустым.")
        os.Exit(1)
    }

    fmt.Print("Введите сообщение M2 (отправляется B, внутри шифрования): ")
    m2, _ := reader.ReadString('\n')
    m2 = strings.TrimSpace(m2)
    if m2 == "" {
        fmt.Println("Ошибка: M2 не может быть пустым.")
        os.Exit(1)
    }

    fmt.Print("Введите сообщение M4 (отправляется A, внутри шифрования): ")
    m4, _ := reader.ReadString('\n')
    m4 = strings.TrimSpace(m4)
    if m4 == "" {
        fmt.Println("Ошибка: M4 не может быть пустым.")
        os.Exit(1)
    }

    fmt.Print("Введите сообщение M5 (отправляется A, в открытом виде): ")
    m5, _ := reader.ReadString('\n')
    m5 = strings.TrimSpace(m5)
    if m5 == "" {
        fmt.Println("Ошибка: M5 не может быть пустым.")
        os.Exit(1)
    }

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
    fmt.Print("Введите сообщение M3 (отправляется B, в открытом виде): ")
    m3, _ := reader.ReadString('\n')
    m3 = strings.TrimSpace(m3)
    if m3 == "" {
        fmt.Println("Ошибка: M3 не может быть пустым.")
        os.Exit(1)
    }

    go func() {
        if !RunUserB(m2, m3) {
            println("Сторона B завершилась с ошибкой.")
            os.Exit(1)
        }
    }()

    time.Sleep(1 * time.Second)

    if !RunUserA(m1, m2, m4, m5, useTimestamp) {
        println("Сторона A завершилась с ошибкой.")
        os.Exit(1)
    }
}