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
}

func NewUserA(key []byte) *UserA {
    return &UserA{key: key}
}

func (a *UserA) CreateFirstMessage(useTimestamp bool, m1, m2 string) (*Message, error) {
    println("\n--- Сторона A (Инициатор): Подготовка первого сообщения ---")
    identifier := GenerateIdentifier(useTimestamp)
    a.originalIdentifier = identifier
    idType := "Timestamp"
    if !useTimestamp {
        idType = "Random number"
    }
    fmt.Printf("A сгенерировал идентификатор (%s): %s\n", idType, identifier)

    payloadBytes, err := SerializePayload(identifier, "A", m1)
    if err != nil {
        println("A не смог сериализовать данные:", err.Error())
        return nil, err
    }

    iv, ciphertext := EncryptAES(a.key, payloadBytes)
    if iv == nil || ciphertext == nil {
        println("A не смог зашифровать данные")
        return nil, fmt.Errorf("encryption failed")
    }

   
    msg := &Message{
        IV:         iv,
        Ciphertext: ciphertext,
        M:          m2,
    }
    return msg, nil
}

func (a *UserA) VerifySecondMessage(received *Message) bool {
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
        println("A не смог расшифровать сообщение B. Идентификация не удалась (вероятно, неверный ключ).")
        return false
    }

    payload, err := DeserializePayload(decryptedBytes)
    if err != nil {
        println("A не смог десериализовать данные от B. Идентификация не удалась.")
        return false
    }

    fmt.Printf("A расшифровал данные: Идентификатор='%s', Отправитель='%s', M3='%s'\n",
        payload.ID, payload.Sender, payload.Message)
    fmt.Printf("A получил M4 (в открытом виде): %s\n", received.M)

    fmt.Println("\nПроверка A:")
    fmt.Printf("  Полученный отправитель: '%s' == 'B'?\n", payload.Sender)
    fmt.Printf("  Полученный идентификатор: '%s' == Оригинальный идентификатор: '%s'?\n",
        payload.ID, a.originalIdentifier)

    if payload.Sender == "B" && payload.ID == a.originalIdentifier {
        println("\n--- ИДЕНТИФИКАЦИЯ УСПЕШНА ---")
        println("Сторона A успешно идентифицировала сторону B.")
        return true
    }

    println("\n--- ИДЕНТИФИКАЦИЯ НЕ УДАЛАСЬ ---")
    if payload.Sender != "B" {
        println("Причина: Полученный ID отправителя не 'B'.")
    }
    if payload.ID != a.originalIdentifier {
        println("Причина: Полученный идентификатор не совпадает с отправленным.")
    }
    return false
}

func RunUserA(m1, m2 string, useTimestamp bool) bool {
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

    msg, err := userA.CreateFirstMessage(useTimestamp, m1, m2)
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

    return userA.VerifySecondMessage(&received)
}