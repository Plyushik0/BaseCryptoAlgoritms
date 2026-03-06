package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"time"
	"sample-app/cipher"
)

func main() {
	clientPubPath := "fiat_pub.json"
	serverPrivPath := "fiat_server_priv.json"
	serverPubPath := "fiat_server_pub.json"


	if _, err := os.Stat(serverPrivPath); os.IsNotExist(err) {
		fmt.Println("Генерируем ключи сервера...")
		cipher.GenerateFiatKeys(32, 128, serverPubPath, serverPrivPath)
	}


	if _, err := os.Stat(serverPrivPath); os.IsNotExist(err) {
		fmt.Println("Нет приватного ключа сервера! Сначала сгенерируйте ключи сервера.")
		os.Exit(1)
	}
	if _, err := os.Stat(clientPubPath); os.IsNotExist(err) {
		fmt.Println("Нет публичного ключа клиента! Сначала сгенерируйте ключи клиента.")
		os.Exit(1)
	}


	Bclient, nclient, err := LoadFiatShamirPubKey(clientPubPath)
	if err != nil {
		fmt.Println("Ошибка загрузки публичного ключа клиента:", err)
		os.Exit(1)
	}
	Bserver, nserver, err := LoadFiatShamirPubKey(serverPubPath)
	if err != nil {
		fmt.Println("Ошибка загрузки публичного ключа сервера:", err)
		os.Exit(1)
	}
	Aserver, _, _, npriv, err := LoadFiatShamirPrivKey(serverPrivPath)
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
		go handleConn(conn, Bclient, nclient, Bserver, nserver, Aserver, npriv)
	}
}

func handleConn(conn net.Conn, Bclient []*big.Int, nclient *big.Int, Bserver []*big.Int, nserver *big.Int, Aserver []*big.Int, npriv *big.Int) {

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

	// --- Десериализация запроса на штамп времени ---
	var tsReq TimeStampRequest
	if err := json.Unmarshal(msg, &tsReq); err != nil {
		fmt.Println("Ошибка парсинга запроса:", err)
		return
	}
	fmt.Printf("Сервер: len(Bclient)=%d, nclient=%s\n", len(Bclient), nclient.String())
	fmt.Printf("s=%x, t=%x, u=%x\n", tsReq.S, tsReq.T, tsReq.U)
	fmt.Println("hashAlg:", tsReq.HashAlg)


	msgToCheck := append([]byte("data"), tsReq.U.Bytes()...)
	valid := verifySignatureFiat(msgToCheck, tsReq.S, tsReq.T, tsReq.U, Bclient, nclient, tsReq.HashAlg)
	fmt.Println("Результат проверки подписи клиента:", valid)
	fmt.Printf("Хеш сообщения: %x\n", tsReq.Hash)
	fmt.Printf("Подпись клиента: s=%x, t=%x, u=%x\n", tsReq.S, tsReq.T, tsReq.U)
	if !valid {
		resp := map[string]interface{}{
			"status":  "ошибка",
			"message": "некорректная подпись",
		}
		sendResponse(conn, resp)
		return
	}

	// --- Формирование временной метки и подписи центра ---
	timestamp := time.Now().Unix()
	dataToSign := append(tsReq.Hash, []byte(time.Unix(timestamp, 0).UTC().Format("060102150405Z"))...)
	_, serverS, serverT, serverU, _, err := signDataForTimeStampFiat(
		dataToSign, Aserver, npriv, tsReq.HashAlg,
	)
	if err != nil {
		fmt.Println("Ошибка подписи центра:", err)
		return
	}

	// --- Формирование и отправка ответа ---
	tsResp := TimeStampResponse{
		Timestamp:  timestamp,
		ServerS:    serverS,
		ServerT:    serverT,
		ServerU:    serverU,
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
