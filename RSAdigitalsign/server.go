package main

import (
    "encoding/json"
    "fmt"
    "io"
    "math/big"
    "net"
    "os"
    "sample-app/cipher"

)

func LoadPubKey(path string) (*big.Int, *big.Int, error) {
    var file map[string]interface{}
    if err := cipher.LoadFromFile(path, &file); err != nil {
        return nil, nil, err
    }
    info := file["SubjectPublicKeyInfo"].(map[string]interface{})
    e := new(big.Int)
    n := new(big.Int)
    e.SetString(info["publicExponent"].(string), 10)
    n.SetString(info["N"].(string), 10)
    return e, n, nil
}

func LoadPrivKey(path string) (*big.Int, *big.Int, error) {
    var file map[string]interface{}
    if err := cipher.LoadFromFile(path, &file); err != nil {
        return nil, nil, err
    }
    d := new(big.Int)
    d.SetString(file["privateExponent"].(string), 10)

    _, nVal, err := LoadPubKey("server_pub.json")
    if err == nil {
        return d, nVal, nil
    }
    return d, nil, err
}

func main() {
    clientPubPath := "client_pub.json"
    serverPrivPath := "server_priv.json"
    serverPubPath := "server_pub.json"

    if _, err := os.Stat(serverPrivPath); os.IsNotExist(err) {
        fmt.Println("Генерируем ключи сервера...")
        cipher.GenerateRSAKeys(1024, serverPubPath, serverPrivPath)
    }
    if _, err := os.Stat(clientPubPath); os.IsNotExist(err) {
        fmt.Println("Нет публичного ключа клиента! Сначала сгенерируйте ключи клиента.")
        os.Exit(1)
    }

    clientE, clientN, err := LoadPubKey(clientPubPath)
    if err != nil {
        fmt.Println("Ошибка загрузки публичного ключа клиента:", err)
        os.Exit(1)
    }
    serverD, serverN, err := LoadPrivKey(serverPrivPath)
    if err != nil {
        fmt.Println("Ошибка загрузки приватного ключа сервера:", err)
        os.Exit(1)
    }

    addr := "127.0.0.1:55556"
    ln, err := net.Listen("tcp", addr)
    if err != nil {
        fmt.Println("Ошибка запуска сервера:", err)
        return
    }
    fmt.Println("Сервер слушает на", addr)

    for {
        conn, err := ln.Accept()
        if err != nil {
            fmt.Println("Ошибка соединения:", err)
            continue
        }
        go handleConn(conn, clientE, clientN, serverD, serverN)
    }
}

func handleConn(conn net.Conn, clientE, clientN, serverD, serverN *big.Int) {
    defer conn.Close()

    var lengthBuf [4]byte
    if _, err := io.ReadFull(conn, lengthBuf[:]); err != nil {
        fmt.Println("Ошибка чтения длины:", err)
        return
    }
    length := int(uint32(lengthBuf[0])<<24 | uint32(lengthBuf[1])<<16 | uint32(lengthBuf[2])<<8 | uint32(lengthBuf[3]))

    msg := make([]byte, length)
    if _, err := io.ReadFull(conn, msg); err != nil {
        fmt.Println("Ошибка чтения сообщения:", err)
        return
    }

   
    var tsReq TimeStampRequest
    if err := json.Unmarshal(msg, &tsReq); err != nil {
        fmt.Println("Ошибка парсинга запроса:", err)
        return
    }

    
    valid := verifySignature(tsReq.Hash, tsReq.Signature, clientE, clientN)
    fmt.Println("Результат проверки подписи клиента:", valid)
    fmt.Printf("Хеш сообщения: %x\n", tsReq.Hash)
    if len(tsReq.Signature) > 0 {
        fmt.Printf("Подпись клиента: %x\n", tsReq.Signature)
    }
    if !valid {
        resp := map[string]interface{}{
            "status":  "ошибка",
            "message": "некорректная подпись",
        }
        sendResponse(conn, resp)
        return
    }

    
    serverSignature, timestamp := serverSign(tsReq.Hash, tsReq.HashAlg, serverD, serverN)
    

   
    tsResp := TimeStampResponse{
        Timestamp:  timestamp,
        ServerSign: serverSignature,
        ServerCert: "server_cert",
    }
    respBytes, _ := json.Marshal(tsResp)

    lengthResp := uint32(len(respBytes))
    lengthBufResp := []byte{
        byte(lengthResp >> 24), byte(lengthResp >> 16), byte(lengthResp >> 8), byte(lengthResp),
    }
    conn.Write(lengthBufResp)
    conn.Write(respBytes)
}

func sendResponse(conn net.Conn, resp map[string]interface{}) {
    respBytes, _ := json.Marshal(resp)
    length := uint32(len(respBytes))
    lengthBuf := []byte{
        byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length),
    }
    conn.Write(lengthBuf)
    conn.Write(respBytes)
}
