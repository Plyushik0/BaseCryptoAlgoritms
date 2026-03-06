package main

import (
    "encoding/binary"
    "encoding/json"
    "math/rand"
    "time"
)

type Info struct {
    ID      string `json:"id"`
    Sender  string `json:"sender"`
    Message string `json:"message"`
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
    randomNum := rand.Int63()
    buf := make([]byte, 8)
    binary.BigEndian.PutUint64(buf, uint64(randomNum))
    return string(buf)
}

func SerializePayload(id, sender, message string) ([]byte, error) {
    payload := Info{ID: id, Sender: sender, Message: message}
    return json.Marshal(payload)
}

func DeserializePayload(data []byte) (*Info, error) {
    var payload Info
    if err := json.Unmarshal(data, &payload); err != nil {
        return nil, err
    }
    return &payload, nil
}