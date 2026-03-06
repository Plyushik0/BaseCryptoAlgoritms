package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"

	"sample-app/cipher"
)




func LoadElGamalPubKey(path string) (p, alpha, beta *big.Int, err error) {
	var file map[string]interface{}
	if err = cipher.LoadFromFile(path, &file); err != nil {
		return
	}
	info := file["SubjectPublicKeyInfo"].(map[string]interface{})
	p = new(big.Int)
	alpha = new(big.Int)
	beta = new(big.Int)
	p.SetString(info["p"].(string), 10)
	alpha.SetString(info["alpha"].(string), 10)
	beta.SetString(info["beta"].(string), 10)
	return
}


func LoadElGamalPrivKey(path string) (a *big.Int, err error) {
	var file map[string]interface{}
	if err = cipher.LoadFromFile(path, &file); err != nil {
		return
	}
	a = new(big.Int)
	a.SetString(file["privateExponent"].(string), 10)
	return
}

func main() {
	clientPrivPath := "elg_priv.json"
	clientPubPath := "elg_pub.json"
	serverPubPath := "elg_server_pub.json"

	
	if _, err := os.Stat(clientPrivPath); os.IsNotExist(err) {
		fmt.Println("Генерируем ключи клиента...")
		cipher.GenerateELGKeys(256, clientPubPath, clientPrivPath)
	}

	
	if _, err := os.Stat(clientPrivPath); os.IsNotExist(err) {
		fmt.Println("Нет приватного ключа клиента! Сначала сгенерируйте ключи.")
		os.Exit(1)
	}
	if _, err := os.Stat(serverPubPath); os.IsNotExist(err) {
		fmt.Println("Нет публичного ключа сервера! Сначала сгенерируйте ключи сервера.")
		os.Exit(1)
	}


	a, err := LoadElGamalPrivKey(clientPrivPath)
	if err != nil {
		fmt.Println("Ошибка загрузки приватного ключа клиента:", err)
		os.Exit(1)
	}
	p, alpha, _, err := LoadElGamalPubKey(clientPubPath)
	if err != nil {
		fmt.Println("Ошибка загрузки публичного ключа клиента:", err)
		os.Exit(1)
	}
	serverP, serverAlpha, serverBeta, err := LoadElGamalPubKey(serverPubPath)
	if err != nil {
		fmt.Println("Ошибка загрузки публичного ключа сервера:", err)
		os.Exit(1)
	}


	data := []byte("data")
	hash, gamma, delta, hashAlg, signErr := signDataForTimeStampElGamal(
		data, p, alpha, a, "streebog256",
	)
	if signErr != nil {
		fmt.Println("Ошибка подписи клиента:", signErr)
		os.Exit(1)
	}

	
	tsReq := TimeStampRequest{
		Hash:    hash,
		Gamma:   gamma,
		Delta:   delta,
		HashAlg: hashAlg,
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

	
	ok := verifyCenterSignatureElGamal(
		hash, tsResp.Timestamp, tsResp.ServerGamma, tsResp.ServerDelta, hashAlg, serverP, serverAlpha, serverBeta,
	)
	fmt.Println("Проверка подписи центра:", ok)


	fmt.Println("Итоговая структура")
	fmt.Printf("Хеш сообщения: %x\n", hash)
	fmt.Printf("Подпись клиента: gamma=%x, delta=%x\n", gamma, delta)
	fmt.Printf("Временная метка: %s\n", time.Unix(tsResp.Timestamp, 0).UTC().Format("060102150405Z"))
	fmt.Printf("Подпись центра: gamma=%x, delta=%x\n", tsResp.ServerGamma, tsResp.ServerDelta)
	fmt.Printf("Сертификат центра: %s\n", tsResp.ServerCert)
}
