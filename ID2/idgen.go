package main

import (
    "encoding/binary"
    "encoding/json"
    "time"
    "crypto/rand"
    "encoding/hex"
    mathrand "math/rand"
)

type Payload struct {
    ID      string `json:"id"`
    Sender  string `json:"sender"`
    Message string `json:"message"`
    R       string `json:"r"`       // Основное случайное число (r_A или r_B)
    ROther  string `json:"r_other"` // Второе случайное число (r_A или r_B) для проверки
    Data    string `json:"data"`    // Вложенные зашифрованные данные (r_A, r_B, Sender, M)
}

type Message struct {
    IV         []byte `json:"iv"`
    Ciphertext []byte `json:"ciphertext"`
    M          string `json:"m"` 
}

func GenerateIdentifier(useTimestamp bool) string {
    if useTimestamp {
        return time.Now().UTC().Format("060102150405Z")
    }
    randomNum := mathrand.Int63()
    buf := make([]byte, 8)
    binary.BigEndian.PutUint64(buf, uint64(randomNum))
    return string(buf)
}

func GenerateRandomNumber() string {
    b := make([]byte, 8)
    if _, err := rand.Read(b); err != nil {
        return ""
    }
    return hex.EncodeToString(b)
}
func SerializePayload(id, sender, message, r, rOther, data string) ([]byte, error) {
    payload := Payload{ID: id, Sender: sender, Message: message, R: r, ROther: rOther, Data: data}
    return json.Marshal(payload)
}

func DeserializePayload(data []byte) (*Payload, error) {
    var payload Payload
    if err := json.Unmarshal(data, &payload); err != nil {
        return nil, err
    }
    return &payload, nil
}