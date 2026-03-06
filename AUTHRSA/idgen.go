package main

import (
    "encoding/json"
    "crypto/rand"
    "encoding/hex"
    "time"
    "math/big"
    "crypto/sha256"
)


type Payload struct {
    ID      string `json:"id"`
    Sender  string `json:"sender"`
    R       string `json:"r"`       
    ROther  string `json:"r_other"` 
}


type FirstMessage struct {
    HashZ     string      `json:"hash_z"`    
    A         string      `json:"a"`         
    Ciphertext []*big.Int `json:"ciphertext"`
}


func GenerateIdentifier(useTimestamp bool) string {
    if useTimestamp {
        return time.Now().UTC().Format("060102150405Z")
    }
    b := make([]byte, 8)
    if _, err := rand.Read(b); err != nil {
        return ""
    }
    return hex.EncodeToString(b)
}


func GenerateRandomNumber() string {
    b := make([]byte, 8)
    if _, err := rand.Read(b); err != nil {
        return ""
    }
    return hex.EncodeToString(b)
}


func HashString(s string) string {
    h := sha256.Sum256([]byte(s))
    return hex.EncodeToString(h[:])
}


func SerializePayload(id, sender, r, rOther string) ([]byte, error) {
    payload := Payload{ID: id, Sender: sender, R: r, ROther: rOther}
    return json.Marshal(payload)
}


func DeserializePayload(data []byte) (*Payload, error) {
    var payload Payload
    if err := json.Unmarshal(data, &payload); err != nil {
        return nil, err
    }
    return &payload, nil
}