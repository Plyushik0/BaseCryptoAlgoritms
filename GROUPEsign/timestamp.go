package main

import (
    "encoding/json"
    "fmt"
    "math/big"
    "net"
    "time"
    "sample-app/mylib"
    "encoding/hex"
)

func startTimestampServer() error {
    var tsPrivateKey TSPrivateKey
    if err := loadJSON("ts_private_key.json", &tsPrivateKey); err != nil {
        return fmt.Errorf("ошибка загрузки ключей TS: %v", err)
    }
    var publicParams PublicParams
    if err := loadJSON("group_public_params.json", &publicParams); err != nil {
        return fmt.Errorf("ошибка загрузки публичных параметров: %v", err)
    }

    listener, err := net.Listen("tcp", "localhost:9999")
    if err != nil {
        return fmt.Errorf("ошибка запуска TS: %v", err)
    }
    defer listener.Close()
    fmt.Println("Сервер временных меток запущен на localhost:9999")

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Printf("Ошибка принятия соединения: %v\n", err)
            continue
        }
        go handleTSConnection(conn, tsPrivateKey, publicParams.TSPublicKey)
    }
}

func handleTSConnection(conn net.Conn, tsPrivateKey TSPrivateKey, tsPublicKey TSPubKey) {
    defer conn.Close()
    data := make([]byte, 4096)
    n, err := conn.Read(data)
    if err != nil {
        fmt.Printf("Ошибка чтения данных TS: %v\n", err)
        conn.Write([]byte(`{"error":"Ошибка чтения данных"}`))
        return
    }

    var request map[string]string
    if err := json.Unmarshal(data[:n], &request); err != nil {
        fmt.Printf("Ошибка декодирования JSON TS: %v\n", err)
        conn.Write([]byte(`{"error":"Неверный JSON"}`))
        return
    }
    fmt.Printf("TS запрос: %v\n", request)

    sigEncoded := request["group_signature_encoded"]
    if sigEncoded == "" {
        fmt.Printf("Ошибка TS: отсутствует подпись\n")
        conn.Write([]byte(`{"error":"Отсутствует подпись"}`))
        return
    }
    sigBytes, err := hex.DecodeString(sigEncoded)
    if err != nil { // Оставляем без детальной обработки ошибок
        fmt.Printf("Ошибка TS: неверный hex формат подписи\n")
        conn.Write([]byte(`{"error":"Неверный hex формат подписи"}`))
        return
    }

    timestamp := time.Now().UTC().Format("20060102150405Z")
    tsData := append(sigBytes, []byte(timestamp)...)
    tsHash := hashMessage(tsData)
    nBig := new(big.Int).Mul(tsPrivateKey.P, tsPrivateKey.Q)
    tsSignature := mylib.FastPowMod(tsHash, tsPrivateKey.D, nBig)

    response := map[string]interface{}{
        "timestamp":     timestamp,
        "ts_signature":  tsSignature.String(),
        "ts_public_key": map[string]interface{}{
            "n": nBig.String(),
            "e": tsPublicKey.E.String(),
        },
    }
    fmt.Printf("TS ответ: %v\n", response)
    if err := json.NewEncoder(conn).Encode(response); err != nil {
        fmt.Printf("Ошибка отправки TS ответа: %v\n", err)
    }
    fmt.Println("Отправлена временная метка.")
}

