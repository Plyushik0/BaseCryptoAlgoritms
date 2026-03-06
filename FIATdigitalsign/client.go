package main

import (
	"encoding/json"
	"fmt"

	"net"
	"os"
	"time"
	"sample-app/cipher"
)

func main() {
	pubPath := "fiat_pub.json"
	privPath := "fiat_priv.json"
	serverPubPath := "fiat_server_pub.json"

	
	if _, err := os.Stat(privPath); os.IsNotExist(err) {
		fmt.Println("Генерируем ключи клиента...")
		cipher.GenerateFiatKeys(32, 128, pubPath, privPath)
	}


	if _, err := os.Stat(privPath); os.IsNotExist(err) {
		fmt.Println("Нет приватного ключа клиента! Сначала сгенерируйте ключи.")
		os.Exit(1)
	}
	if _, err := os.Stat(serverPubPath); os.IsNotExist(err) {
		fmt.Println("Нет публичного ключа сервера! Сначала сгенерируйте ключи сервера.")
		os.Exit(1)
	}

	
	A, _, _, n, err := LoadFiatShamirPrivKey(privPath)
	if err != nil {
		fmt.Println("Ошибка загрузки приватного ключа клиента:", err)
		os.Exit(1)
	}
	Bserver, nserver, err := LoadFiatShamirPubKey(serverPubPath)
	if err != nil {
		fmt.Println("Ошибка загрузки публичного ключа сервера:", err)
		os.Exit(1)
	}

	data := []byte("data")
	hash, s, t, u, hashAlg, signErr := signDataForTimeStampFiat(
		data, A, n, "streebog256",
	)
	if signErr != nil {
		fmt.Println("Ошибка подписи клиента:", signErr)
		os.Exit(1)
	}
	fmt.Printf("Клиент: len(A)=%d, n=%s\n", len(A), n.String())
	fmt.Printf("s=%x, t=%x, u=%x\n", s, t, u)
	fmt.Println("hashAlg:", hashAlg)

	fmt.Printf("u (client) = %x\n", u.Bytes())
	fmt.Printf("h (client) = %x\n", hash)
	fmt.Printf("s (client) = %x\n", s)

	tsReq := TimeStampRequest{
		Hash:    hash,
		S:       s,
		T:       t,
		U:       u,
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
	if tsResp.ServerT == nil || tsResp.ServerU == nil || tsResp.ServerS == nil || tsResp.Timestamp == 0 {
		var errResp map[string]interface{}
		if err := json.Unmarshal(respBuf, &errResp); err == nil {
			fmt.Printf("Ошибка от сервера: %v\n", errResp)
		} else {
			fmt.Println("Ошибка: сервер не вернул подпись центра (t, u, s).")
			fmt.Printf("Ответ сервера: %+v\n", tsResp)
		}
		return
	}


	dataToCheck := append(hash, []byte(time.Unix(tsResp.Timestamp, 0).UTC().Format("060102150405Z"))...)
	fmt.Printf("dataToCheck (клиент) = %x\n", dataToCheck)
	ok := verifySignatureFiat(
    	dataToCheck, tsResp.ServerS, tsResp.ServerT, tsResp.ServerU, Bserver, nserver, hashAlg,
	)
	fmt.Println("Проверка подписи центра:", !ok)


	fmt.Println("Итоговая структура")
	fmt.Printf("Хеш сообщения: %x\n", hash)
	fmt.Printf("Подпись клиента: s=%x, t=%x, u=%x\n", s, t, u)
	fmt.Printf("Временная метка: %s\n", time.Unix(tsResp.Timestamp, 0).UTC().Format("060102150405Z"))
	fmt.Printf("Подпись центра: s=%x, t=%x, u=%x\n", tsResp.ServerS, tsResp.ServerT, tsResp.ServerU)
	fmt.Printf("Сертификат центра: %s\n", tsResp.ServerCert)
}
