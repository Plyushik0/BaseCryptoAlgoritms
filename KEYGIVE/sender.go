package main

import (
	"crypto/rand"

	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"sample-app/cipher"

	"time"
)

type Sender struct {
	role              string
	publicKey         map[string]*big.Int 
	privateKey        map[string]*big.Int 
	sessionKey        []byte
	receiverPublicKey map[string]*big.Int
}

func NewSender() *Sender {
	return &Sender{role: "Sender A"}
}

func (s *Sender) logMessage(message string) {
	fmt.Printf("[%s] %s\n", s.role, message)
}

func (s *Sender) generateKeys() {
	s.logMessage("Генерация ключей RSA...")
	e, d, n, _ := cipher.GenerateRSAKeys(1024, "", "") 
	s.publicKey = map[string]*big.Int{"e": e, "n": n}
	s.privateKey = map[string]*big.Int{"d": d, "n": n}
	s.logMessage("Ключи RSA сгенерированы.")
}

func (s *Sender) generateSessionKey() {
	s.logMessage("Генерация 256-битного сеансового ключа...")
	s.sessionKey = make([]byte, 32) 
	_, err := rand.Read(s.sessionKey)
	if err != nil {
		s.logMessage(fmt.Sprintf("Ошибка генерации ключа: %v", err))
		return
	}
	s.logMessage(fmt.Sprintf("Сеансовый ключ сгенерирован: %x", s.sessionKey))
}

func (s *Sender) getReceiverPublicKey() bool {
	s.logMessage("Подключение к получателю на localhost:12345 для получения публичного ключа...")
	conn, err := net.Dial("tcp", "localhost:12345") 
	if err != nil {
		s.logMessage(fmt.Sprintf("Ошибка подключения: %v", err))
		return false
	}
	defer conn.Close()

	request := map[string]string{"action": "get_public_key"}
	if !SendJSON(conn, request) {
		s.logMessage("Ошибка отправки запроса на публичный ключ.")
		return false
	}

	keyInterface, err := ReceiveJSON(conn)
	if err != nil {
		s.logMessage(fmt.Sprintf("Ошибка получения ключа: %v", err))
		return false
	}

	keyData, ok := keyInterface.(map[string]interface{})
	if !ok {
		s.logMessage("Полученный ключ не в ожидаемом формате.")
		return false
	}

	nStr, ok := keyData["n"].(string)
	eStr, ok2 := keyData["e"].(string)
	if !ok || !ok2 {
		s.logMessage("Ошибка в формате данных ключа.")
		return false
	}

	n, success := new(big.Int).SetString(nStr, 10)
	e, success2 := new(big.Int).SetString(eStr, 10)
	if !success || !success2 {
		s.logMessage("Не удалось преобразовать ключ в big.Int.")
		return false
	}

	s.receiverPublicKey = map[string]*big.Int{"n": n, "e": e}
	s.logMessage("Публичный ключ получателя получен.")
	return true
}

func (s *Sender) Run() {
	s.generateKeys()
	s.generateSessionKey()

	if !s.getReceiverPublicKey() {
		s.logMessage("Не удалось получить публичный ключ получателя. Завершение.")
		return
	}

	conn, err := net.Dial("tcp", "localhost:12345") 
	if err != nil {
		s.logMessage(fmt.Sprintf("Ошибка подключения к получателю: %v", err))
		return
	}
	defer conn.Close()

	timestamp := time.Now().Unix()
	s.logMessage(fmt.Sprintf("Временная метка: %d", timestamp))

	messageToSign := append(s.sessionKey, []byte(fmt.Sprintf("%d", timestamp))...)
	marshaledKey, err := json.Marshal(map[string]string{
		"e": s.receiverPublicKey["e"].String(),
		"n": s.receiverPublicKey["n"].String(),
	})
	if err != nil {
		s.logMessage(fmt.Sprintf("Ошибка сериализации публичного ключа: %v", err))
		return
	}
	messageToSign = append(messageToSign, marshaledKey...)


	_, signature, usedHashAlg, err := signDataForTimeStamp(messageToSign, s.privateKey["d"], s.privateKey["n"], "sha512")
	if err != nil {
		s.logMessage(fmt.Sprintf("Ошибка создания подписи: %v", err))
		return
	}
	s.logMessage(fmt.Sprintf("Создана подпись с %s: %x", usedHashAlg, signature))

	dataToEncrypt := map[string]interface{}{
		"key":       fmt.Sprintf("%x", s.sessionKey),
		"timestamp": timestamp,
		"signature": fmt.Sprintf("%x", signature),
	}
	dataBytes, err := json.Marshal(dataToEncrypt)
	if err != nil {
		s.logMessage(fmt.Sprintf("Ошибка сериализации данных: %v", err))
		return
	}

	encryptedBlocks := cipher.EncryptRSA(string(dataBytes), s.receiverPublicKey["e"], s.receiverPublicKey["n"])
	s.logMessage("Данные зашифрованы публичным ключом получателя.")

	
	encryptedData := make([]string, len(encryptedBlocks))
	for i, block := range encryptedBlocks {
		encryptedData[i] = block.String()
	}

	messageToSend := map[string]interface{}{
		"action":           "send_key",
		"encrypted_data":   encryptedData,
		"sender_public_key": map[string]string{
			"e": s.publicKey["e"].String(),
			"n": s.publicKey["n"].String(),
		},
	}
	if !SendJSON(conn, messageToSend) {
		s.logMessage("Ошибка отправки зашифрованного ключа.")
		return
	}
	s.logMessage("Зашифрованный ключ отправлен получателю.")

	response, err := ReceiveJSON(conn)
	if err != nil {
		s.logMessage(fmt.Sprintf("Ошибка получения ответа: %v", err))
		return
	}
	s.logMessage(fmt.Sprintf("Ответ от получателя: %v", response))
}

func main() {
	sender := NewSender()
	sender.Run()
}