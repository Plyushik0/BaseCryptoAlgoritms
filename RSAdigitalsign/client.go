package main

import (
    "encoding/json"
    "fmt"
    "math/big"
    "net"
    "os"
    "sample-app/cipher"
    "sample-app/hashf"
    "sample-app/mylib"
    "time"
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

    _, nVal, err := LoadPubKey("client_pub.json")
    if err == nil {
        return d, nVal, nil
    }
    return d, nil, err
}



func main() {
    clientPrivPath := "client_priv.json"
    clientPubPath := "client_pub.json"
    serverPubPath := "server_pub.json"

    if _, err := os.Stat(clientPrivPath); os.IsNotExist(err) {
        fmt.Println("Генерируем ключи клиента...")
        cipher.GenerateRSAKeys(1024, clientPubPath, clientPrivPath)
    }
    if _, err := os.Stat(serverPubPath); os.IsNotExist(err) {
        fmt.Println("Нет публичного ключа сервера! Сначала сгенерируйте ключи сервера.")
        os.Exit(1)
    }

    clientD, clientN, err := LoadPrivKey(clientPrivPath)
    if err != nil {
        fmt.Println("Ошибка загрузки приватного ключа клиента:", err)
        os.Exit(1)
    }
    serverE, serverN, err := LoadPubKey(serverPubPath)
    if err != nil {
        fmt.Println("Ошибка загрузки публичного ключа сервера:", err)
        os.Exit(1)
    }

    
    data := []byte("data")
    hash, sigBytes, hashAlg, err := signDataForTimeStamp(data, clientD, clientN, "streebog256")
    if err != nil {
        fmt.Println("Ошибка подписи данных клиентом:", err)
        os.Exit(1)
    }

    
    tsReq := TimeStampRequest{
        Hash:      hash,
        Signature: sigBytes,
        HashAlg:   hashAlg,
    }
    message, err := json.Marshal(tsReq)
    if err != nil {
        fmt.Println("Ошибка сериализации запроса:", err)
        os.Exit(1)
    }

    
    conn, err := net.Dial("tcp", "127.0.0.1:55556")
    if err != nil {
        fmt.Println("Ошибка соединения:", err)
        return
    }
    defer conn.Close()

    length := uint32(len(message))
    lengthBuf := []byte{
        byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length),
    }
    conn.Write(lengthBuf)
    conn.Write(message)


    respLenBuf := make([]byte, 4)
    _, err = conn.Read(respLenBuf)
    if err != nil {
        fmt.Println("Ошибка чтения длины ответа:", err)
        return
    }
    respLen := int(uint32(respLenBuf[0])<<24 | uint32(respLenBuf[1])<<16 | uint32(respLenBuf[2])<<8 | uint32(respLenBuf[3]))
    respBuf := make([]byte, respLen)
    _, err = conn.Read(respBuf)
    if err != nil {
        fmt.Println("Ошибка чтения ответа:", err)
        return
    }


    var tsResp TimeStampResponse
    if err := json.Unmarshal(respBuf, &tsResp); err != nil {
        fmt.Println("Ошибка парсинга ответа центра:", err)
        return
    }


    dataToCheck := append(hash, []byte(time.Unix(tsResp.Timestamp, 0).UTC().Format("060102150405Z"))...)
    var hashToCheck []byte
    switch hashAlg {
    case "SHA256":
        h := hashf.Sha256(dataToCheck)
        hashToCheck = h[:]
    case "GOST256":
        hashToCheck = hashf.StreebogHash(dataToCheck, 32)
    default:
        h := hashf.Sha256(dataToCheck)
        hashToCheck = h[:]
    }
    m := new(big.Int).SetBytes(tsResp.ServerSign)
    hashRecovered := mylib.FastPowMod(m, serverE, serverN).Bytes()
    ok := string(hashRecovered) == string(hashToCheck)
    fmt.Println("Проверка подписи сервера:", ok)


    fmt.Printf("Хеш сообщения: %x\n", hash)
    fmt.Printf("Подпись клиента: %x\n", sigBytes)
    fmt.Printf("Временная метка: %s\n", time.Unix(tsResp.Timestamp, 0).UTC().Format("060102150405Z"))
    fmt.Printf("Подпись центра: %x\n", tsResp.ServerSign)
    fmt.Printf("Сертификат центра: %s\n", tsResp.ServerCert)
}