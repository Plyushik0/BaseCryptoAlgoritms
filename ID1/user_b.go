package main

import (
    "fmt"
    "net"
    "os"
    "encoding/json"
)

type UserB struct {
    key []byte
}

func NewUserB(key []byte) *UserB {
    return &UserB{key: key}
}

func (b *UserB) ProcessFirstMessageAndCreateResponse(received *Message, m3, m4 string) (*Message, error) {
    println("\n--- Сторона B (Ответчик): Обработка первого сообщения ---")

    
    fmt.Printf("Полученное сообщение от A: %+v\n", received)

  
    ivBytes := make([]byte, len(received.IV))
    for i, v := range received.IV {
        ivBytes[i] = byte(v)
    }
    cipherBytes := make([]byte, len(received.Ciphertext))
    for i, v := range received.Ciphertext {
        cipherBytes[i] = byte(v)
    }

    decryptedBytes := DecryptAES(b.key, ivBytes, cipherBytes)
    if decryptedBytes == nil {
        println("B не смог расшифровать сообщение A. Проверка не удалась.")
        return nil, fmt.Errorf("decryption failed")
    }

    payload, err := DeserializePayload(decryptedBytes)
    if err != nil {
        println("B не смог десериализовать данные от A. Проверка не удалась.")
        return nil, err
    }

    fmt.Printf("B расшифровал данные: Идентификатор='%s', Отправитель='%s', M1='%s'\n",
        payload.ID, payload.Sender, payload.Message)
    fmt.Printf("B получил M2 (в открытом виде): %s\n", received.M)

    fmt.Printf("\nПроверка B: Полученный отправитель: '%s' == 'A'?\n", payload.Sender)
    if payload.Sender != "A" {
        println("Проверка не удалась: Полученный ID отправителя не 'A'.")
        println("\n--- ИДЕНТИФИКАЦИЯ НЕ УДАЛАСЬ (B не смог проверить A) ---")
        return nil, fmt.Errorf("invalid sender")
    }

    println("Проверка A успешна (ID отправителя совпадает).")

    payloadBytes, err := SerializePayload(payload.ID, "B", m3)
    if err != nil {
        println("B не смог сериализовать ответные данные.")
        return nil, err
    }

    iv, ciphertext := EncryptAES(b.key, payloadBytes)
    if iv == nil || ciphertext == nil {
        println("B не смог зашифровать ответные данные.")
        return nil, fmt.Errorf("encryption failed")
    }

    println("\n--- Сторона B (Ответчик): Проверка A успешна, подготовка ответа ---")
    return &Message{IV: iv, Ciphertext: ciphertext, M: m4}, nil
}

func RunUserB(m3, m4 string) bool {
    key, err := os.ReadFile(KEY_FILE)
    if err != nil {
        println("Ошибка загрузки ключа:", err.Error())
        return false
    }
    if len(key) != KEY_SIZE {
        fmt.Printf("Ошибка: Неверный размер ключа (%d байт). Ожидается %d байт.\n", len(key), KEY_SIZE)
        return false
    }

    userB := NewUserB(key)

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

    println("\nОжидание первого сообщения от A...")
    receivedRaw, err := ReceiveJSON(conn)
    if err != nil {
        println("Не удалось получить первое сообщение:", err.Error())
        return false
    }
    println("Первое сообщение получено.")

    
    receivedBytes, err := json.Marshal(receivedRaw)
    if err != nil {
        println("Не удалось сериализовать полученное сообщение:", err.Error())
        return false
    }
    var received Message
    if err := json.Unmarshal(receivedBytes, &received); err != nil {
        println("Не удалось десериализовать полученное сообщение:", err.Error())
        return false
    }

    msg, err := userB.ProcessFirstMessageAndCreateResponse(&received, m3, m4)
    if err != nil {
        println("Не удалось обработать сообщение или создать ответ.")
        return false
    }

    println("\nОтправка второго сообщения A...")
    if !SendJSON(conn, msg) {
        println("Не удалось отправить второе сообщение.")
        return false
    }
    println("Второе сообщение отправлено.")

    println("\n--- Сторона B (Ответчик): Завершил шаг протокола. Проверьте результат на стороне A. ---")
    return true
}