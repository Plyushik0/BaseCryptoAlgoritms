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
	clientPubPath := "elg_pub.json"
	serverPrivPath := "elg_server_priv.json"
	serverPubPath := "elg_server_pub.json"

	if _, err := os.Stat(serverPrivPath); os.IsNotExist(err) {
		fmt.Println("Генерируем ключи сервера...")
		cipher.GenerateELGKeys(256, serverPubPath, serverPrivPath)
	}

	if _, err := os.Stat(serverPrivPath); os.IsNotExist(err) {
		fmt.Println("Нет приватного ключа сервера! Сначала сгенерируйте ключи сервера.")
		os.Exit(1)
	}
	if _, err := os.Stat(clientPubPath); os.IsNotExist(err) {
		fmt.Println("Нет публичного ключа клиента! Сначала сгенерируйте ключи клиента.")
		os.Exit(1)
	}


	clientP, clientAlpha, clientBeta, err := LoadElGamalPubKey(clientPubPath)
	if err != nil {
		fmt.Println("Ошибка загрузки публичного ключа клиента:", err)
		os.Exit(1)
	}
	serverP, serverAlpha, _, err := LoadElGamalPubKey(serverPubPath)
	if err != nil {
		fmt.Println("Ошибка загрузки публичного ключа сервера:", err)
		os.Exit(1)
	}
	serverA, err := LoadElGamalPrivKey(serverPrivPath)
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
		go handleConn(conn, clientP, clientAlpha, clientBeta, serverP, serverAlpha, serverA)
	}
}

func handleConn(conn net.Conn, clientP, clientAlpha, clientBeta, serverP, serverAlpha, serverA *big.Int) {
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

	
	valid := verifySignatureElGamal(tsReq.Hash, tsReq.Gamma, tsReq.Delta, clientP, clientAlpha, clientBeta)
	fmt.Println("Результат проверки подписи клиента:", valid)
	fmt.Printf("Хеш сообщения: %x\n", tsReq.Hash)
	if len(tsReq.Gamma) > 0 && len(tsReq.Delta) > 0 {
		fmt.Printf("Подпись клиента: gamma=%x, delta=%x\n", tsReq.Gamma, tsReq.Delta)
	}
	if !valid {
		resp := map[string]interface{}{
			"status":  "ошибка",
			"message": "некорректная подпись",
		}
		sendResponse(conn, resp)
		return
	}

	
	serverGamma, serverDelta, timestamp, err := serverSignElGamal(tsReq.Hash, tsReq.HashAlg, serverP, serverAlpha, serverA)
	if err != nil {
		fmt.Println("Ошибка подписи центра:", err)
		return
	}
	
	tsResp := TimeStampResponse{
		Timestamp:   timestamp,
		ServerGamma: serverGamma,
		ServerDelta: serverDelta,
		ServerCert:  "server_cert",
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
