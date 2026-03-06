package main

import (
    "fmt"
    "net"
    "os"
    "encoding/json"
)

type UserA struct {
    key              []byte
    originalIdentifier string
    originalR        string
}

func NewUserA(key []byte) *UserA {
    return &UserA{key: key}
}

func (a *UserA) CreateFirstMessage(useTimestamp bool, m1 string) (*Message, error) {
    println("\n--- Сторона A (Инициатор): Подготовка первого сообщения ---")
    identifier := GenerateIdentifier(useTimestamp)
    a.originalIdentifier = identifier
    rA := GenerateRandomNumber()
    a.originalR = rA
    idType := "Timestamp"
    if !useTimestamp {
        idType = "Random number"
    }
    fmt.Printf("A сгенерировал идентификатор (%s): %s, R_A: %s\n", idType, identifier, rA)

    payloadBytes, err := SerializePayload(identifier, "A", m1, rA, "", "")
    if err != nil {
        println("A не смог сериализовать данные:", err.Error())
        return nil, err
    }

    iv, ciphertext := EncryptAES(a.key, payloadBytes)
    if iv == nil || ciphertext == nil {
        println("A не смог зашифровать данные")
        return nil, fmt.Errorf("encryption failed")
    }

    return &Message{IV: iv, Ciphertext: ciphertext, M: ""}, nil
}

func (a *UserA) VerifySecondMessageAndCreateThird(received *Message, m4, m5 string) (*Message, bool) {
    println("\n--- Сторона A (Инициатор): Проверка второго сообщения ---")

    ivBytes := make([]byte, len(received.IV))
    for i, v := range received.IV {
        ivBytes[i] = byte(v)
    }
    cipherBytes := make([]byte, len(received.Ciphertext))
    for i, v := range received.Ciphertext {
        cipherBytes[i] = byte(v)
    }

    decryptedBytes := DecryptAES(a.key, ivBytes, cipherBytes)
    if decryptedBytes == nil {
        println("A не смог расшифровать сообщение B. Идентификация не удалась.")
        return nil, false
    }

    payload, err := DeserializePayload(decryptedBytes)
    if err != nil {
        println("A не смог десериализовать данные от B. Идентификация не удалась.")
        return nil, false
    }

    fmt.Printf("A расшифровал данные: M3='%s', R_A'='%s', R_B='%s', B='%s', M2='%s'\n",
        payload.Message, payload.R, payload.ROther, payload.Sender, payload.Data)
    fmt.Printf("A получил M3 (в открытом виде): %s\n", received.M)

    fmt.Println("\nПроверка A:")
    fmt.Printf("  Полученный отправитель: '%s' == 'B'?\n", payload.Sender)
    fmt.Printf("  Полученное R_A: '%s' == Оригинальное R_A: '%s'?\n", payload.R, a.originalR)

    if payload.Sender != "B" || payload.R != a.originalR {
        println("\n--- ИДЕНТИФИКАЦИЯ НЕ УДАЛАСЬ ---")
        if payload.Sender != "B" {
            println("Причина: Полученный ID отправителя не 'B'.")
        }
        if payload.R != a.originalR {
            println("Причина: Полученное R_A не совпадает с отправленным R_A.")
        }
        return nil, false
    }

    // Третий проход: {M5, E_kAB(r_B, r_A, A, M4)}
    rB := payload.ROther
    encryptedData, err := SerializePayload(a.originalIdentifier, "A", m4, rB, a.originalR, "")
    if err != nil {
        println("A не смог сериализовать данные для третьего сообщения:", err.Error())
        return nil, false
    }
    iv, ciphertext := EncryptAES(a.key, encryptedData)
    if iv == nil || ciphertext == nil {
        println("A не смог зашифровать данные для третьего сообщения.")
        return nil, false
    }

    println("\n--- Сторона A (Инициатор): Проверка успешна, подготовка третьего сообщения ---")
    return &Message{IV: iv, Ciphertext: ciphertext, M: m5}, true
}

func RunUserA(m1, m2, m4, m5 string, useTimestamp bool) bool {
    key, err := os.ReadFile(KEY_FILE)
    if err != nil {
        println("Ошибка загрузки ключа:", err.Error())
        return false
    }
    if len(key) != KEY_SIZE {
        fmt.Printf("Ошибка: Неверный размер ключа (%d байт). Ожидается %d байт.\n", len(key), KEY_SIZE)
        return false
    }

    userA := NewUserA(key)

    conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", HOST, PORT))
    if err != nil {
        fmt.Printf("Ошибка подключения к %s:%d: %s\n", HOST, PORT, err.Error())
        return false
    }
    defer conn.Close()
    println("Соединение установлено.")

    msg, err := userA.CreateFirstMessage(useTimestamp, m1)
    if err != nil {
        println("Не удалось создать первое сообщение.")
        return false
    }

    println("\nОтправка первого сообщения B...")
    if !SendJSON(conn, msg) {
        println("Не удалось отправить первое сообщение.")
        return false
    }
    println("Первое сообщение отправлено.")

    println("\nОжидание второго сообщения от B...")
    receivedRaw, err := ReceiveJSON(conn)
    if err != nil {
        println("Не удалось получить второе сообщение:", err.Error())
        return false
    }
    println("Второе сообщение получено.")

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

    thirdMsg, success := userA.VerifySecondMessageAndCreateThird(&received, m4, m5)
    if !success {
        return false
    }

    println("\nОтправка третьего сообщения B...")
    if !SendJSON(conn, thirdMsg) {
        println("Не удалось отправить третье сообщение.")
        return false
    }
    println("Третье сообщение отправлено.")

    println("\n--- Сторона A (Инициатор): Завершил протокол. Проверьте результат на стороне B. ---")
    return true
}