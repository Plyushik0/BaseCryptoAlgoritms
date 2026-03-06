package main

import (
    "fmt"
    "net"
    "os"
    "encoding/json"

)

type UserB struct {
    key     []byte
    rB      string
}

func NewUserB(key []byte) *UserB {
    return &UserB{key: key}
}

func (b *UserB) ProcessFirstMessageAndCreateResponse(received *Message, m2, m3 string) (*Message, error) {
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

    fmt.Printf("B расшифровал данные: Идентификатор='%s', Отправитель='%s', M1='%s', R_A='%s'\n",
        payload.ID, payload.Sender, payload.Message, payload.R)

    fmt.Printf("\nПроверка B: Полученный отправитель: '%s' == 'A'?\n", payload.Sender)
    if payload.Sender != "A" {
        println("Проверка не удалась: Полученный ID отправителя не 'A'.")
        println("\n--- ИДЕНТИФИКАЦИЯ НЕ УДАЛАСЬ (B не смог проверить A) ---")
        return nil, fmt.Errorf("invalid sender")
    }

    println("Проверка A успешна (ID отправителя совпадает).")

    rA := payload.R
    b.rB = GenerateRandomNumber()
    encryptedData, err := SerializePayload(payload.ID, "B", m2, rA, b.rB, "")
    if err != nil {
        println("B не смог сериализовать данные для второго сообщения.")
        return nil, err
    }
    iv, ciphertext := EncryptAES(b.key, encryptedData)
    if iv == nil || ciphertext == nil {
        println("B не смог зашифровать данные для второго сообщения.")
        return nil, fmt.Errorf("encryption failed")
    }

    println("\n--- Сторона B (Ответчик): Проверка A успешна, подготовка второго сообщения ---")
    return &Message{IV: iv, Ciphertext: ciphertext, M: m3}, nil
}

func (b *UserB) VerifyThirdMessage(received *Message) bool {
    println("\n--- Сторона B (Ответчик): Проверка третьего сообщения ---")

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
        println("B не смог расшифровать третье сообщение. Идентификация не удалась.")
        return false
    }

    payload, err := DeserializePayload(decryptedBytes)
    if err != nil {
        println("B не смог десериализовать данные третьего сообщения. Идентификация не удалась.")
        return false
    }

    fmt.Printf("B расшифровал данные: M4='%s', R_B'='%s', R_A='%s', A='%s'\n",
        payload.Message, payload.R, payload.ROther, payload.Sender)
    fmt.Printf("B получил M5 (в открытом виде): %s\n", received.M)

    fmt.Println("\nПроверка B:")
    fmt.Printf("  Полученный отправитель: '%s' == 'A'?\n", payload.Sender)
    fmt.Printf("  Полученное R_B: '%s' == Оригинальное R_B: '%s'?\n", payload.R, b.rB)

    if payload.Sender != "A" || payload.R != b.rB {
        println("\n--- ИДЕНТИФИКАЦИЯ НЕ УДАЛАСЬ ---")
        if payload.Sender != "A" {
            println("Причина: Полученный ID отправителя не 'A'.")
        }
        if payload.R != b.rB {
            println("Причина: Полученное R_B не совпадает с отправленным R_B.")
        }
        return false
    }

    println("\n--- ИДЕНТИФИКАЦИЯ УСПЕШНА ---")
    println("Сторона B успешно идентифицировала сторону A.")
    return true
}

func RunUserB(m2, m3 string) bool {
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

    msg, err := userB.ProcessFirstMessageAndCreateResponse(&received, m2, m3)
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

    println("\nОжидание третьего сообщения от A...")
    receivedRaw, err = ReceiveJSON(conn)
    if err != nil {
        println("Не удалось получить третье сообщение:", err.Error())
        return false
    }
    println("Третье сообщение получено.")

    receivedBytes, err = json.Marshal(receivedRaw)
    if err != nil {
        println("Не удалось сериализовать полученное сообщение:", err.Error())
        return false
    }
    if err := json.Unmarshal(receivedBytes, &received); err != nil {
        println("Не удалось десериализовать полученное сообщение:", err.Error())
        return false
    }

    return userB.VerifyThirdMessage(&received)
}