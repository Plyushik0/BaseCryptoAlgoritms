package main

import (
    "fmt"
    "net"
    "sample-app/cipher"
    "math/big"
    "encoding/json"
)

// Структура для сообщения (для совместимости с SendJSON/ReceiveJSON)
type Message struct {
    Ciphertext []*big.Int `json:"ciphertext"`
    M          string     `json:"m"`
}

type UserA struct {
    pubKeyE  *big.Int
    pubKeyN  *big.Int
    privKeyD *big.Int
    privKeyN *big.Int

}

func NewUserA(pubE, pubN, privD, privN *big.Int) *UserA {
    return &UserA{pubKeyE: pubE, pubKeyN: pubN, privKeyD: privD, privKeyN: privN}
}


func (a *UserA) CreateFirstMessage(aID string, pubBE, pubBN *big.Int, useTimestamp bool) (*Message, string, string, error) {
    fmt.Println("\n--- Сторона A (Инициатор): Подготовка первого сообщения ---")
    z := GenerateRandomNumber()
    identifier := GenerateIdentifier(useTimestamp)
    hashZ := HashString(z)
    payload, _ := json.Marshal(map[string]string{"z": z, "a": identifier})
    ciphertext := cipher.EncryptRSA(string(payload), pubBE, pubBN)
    if len(ciphertext) == 0 {
        fmt.Println("A не смог зашифровать данные")
        return nil, "", "", fmt.Errorf("encryption failed")
    }
    return &Message{Ciphertext: ciphertext, M: ""}, z, hashZ, nil
}


func (a *UserA) VerifyResponse(zResp, hashZ string) bool {
    fmt.Println("\n--- Сторона A (Инициатор): Проверка ответа от B ---")
    if HashString(zResp) == hashZ {
        fmt.Println("Аутентификация B успешна!")
        return true
    }
    fmt.Println("Аутентификация B не удалась!")
    return false
}

func RunUserA(aID string, pubBE, pubBN *big.Int, useTimestamp bool) bool {
    userA := NewUserA(nil, nil, nil, nil)

    conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", HOST, PORT))
    if err != nil {
        fmt.Printf("Ошибка подключения к %s:%d: %s\n", HOST, PORT, err.Error())
        return false
    }
    defer conn.Close()
    fmt.Println("Соединение установлено.")


    msg, _, hashZ, err := userA.CreateFirstMessage(aID, pubBE, pubBN, useTimestamp)
    if err != nil {
        fmt.Println("Не удалось создать первое сообщение.")
        return false
    }
    fmt.Println("\nОтправка первого сообщения B...")
    if !SendJSON(conn, msg) {
        fmt.Println("Не удалось отправить первое сообщение.")
        return false
    }
    fmt.Println("Первое сообщение отправлено.")

  
    fmt.Println("\nОжидание ответа от B...")
    var zResp string
    if err := ReceiveJSON(conn, &zResp); err != nil {
        fmt.Println("Ошибка получения ответа от B:", err.Error())
        return false
    }
    fmt.Printf("A получил z от B: %s\n", zResp)

  
    return userA.VerifyResponse(zResp, hashZ)
}