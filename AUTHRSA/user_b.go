package main

import (
    "fmt"
    "net"
    "math/big"
    "encoding/json"
    "sample-app/cipher"
)


func RunUserB(privBD, privBN *big.Int) bool {
    ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", HOST, PORT))
    if err != nil {
        fmt.Printf("Ошибка запуска сервера: %s\n", err.Error())
        return false
    }
    defer ln.Close()
    fmt.Printf("Слушаем подключения на %s:%d...\n", HOST, PORT)

    conn, err := ln.Accept()
    if err != nil {
        println("Ошибка принятия соединения:", err.Error())
        return false
    }
    defer conn.Close()
    println("Подключение принято.")

    var received Message
    if err := ReceiveJSON(conn, &received); err != nil {
        println("Не удалось получить первое сообщение:", err.Error())
        return false
    }
    println("Первое сообщение получено.")

    decrypted := cipher.DecryptRSA(received.Ciphertext, privBD, privBN)
    if decrypted == "" {
        println("B не смог расшифровать сообщение A. Проверка не удалась.")
        return false
    }
    var payload map[string]string
    if err := json.Unmarshal([]byte(decrypted), &payload); err != nil {
        println("B не смог десериализовать данные от A. Проверка не удалась.")
        return false
    }
    z := payload["z"]
    aID := payload["a"]
    fmt.Printf("B расшифровал данные: z='%s', A='%s'\n", z, aID)


    if !SendJSON(conn, z) {
        println("Не удалось отправить z в ответ.")
        return false
    }
    println("Аутентификация A успешна!")
    return true
}