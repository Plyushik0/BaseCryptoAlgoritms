package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"sample-app/cipher"
	"sample-app/hashf"

)

type Receiver struct {
	role       string
	publicKey  map[string]*big.Int
	privateKey map[string]*big.Int
}

func NewReceiver() *Receiver {
	return &Receiver{role: "Receiver B"}
}

func (r *Receiver) logMessage(message string) {
	fmt.Printf("[%s] %s\n", r.role, message)
}

func (r *Receiver) generateKeys() {
	r.logMessage("Генерация ключей RSA...")
	e, d, n, _ := cipher.GenerateRSAKeys(1024, "", "") 
	r.publicKey = map[string]*big.Int{"e": e, "n": n}
	r.privateKey = map[string]*big.Int{"d": d, "n": n}
	r.logMessage("Ключи RSA сгенерированы.")
}

func (r *Receiver) Run() {
	r.generateKeys()

	listener, err := net.Listen("tcp", "localhost:12345") 
	if err != nil {
		r.logMessage(fmt.Sprintf("Ошибка привязки сокета: %v", err))
		return
	}
	defer listener.Close()
	r.logMessage("Слушает на localhost:12345")

	for {
		conn, err := listener.Accept()
		if err != nil {
			r.logMessage(fmt.Sprintf("Ошибка принятия подключения: %v", err))
			continue
		}
		r.logMessage(fmt.Sprintf("Принято подключение от %s", conn.RemoteAddr().String()))

		go r.handleConnection(conn)
	}
}

func (r *Receiver) handleConnection(conn net.Conn) {
	defer conn.Close()

	data, err := ReceiveJSON(conn)
	if err != nil {
		r.logMessage(fmt.Sprintf("Ошибка получения данных: %v", err))
		SendJSON(conn, "Error receiving data")
		return
	}

	message, ok := data.(map[string]interface{})
	if !ok {
		r.logMessage("Неверный формат сообщения.")
		SendJSON(conn, "Invalid message format")
		return
	}

	action, ok := message["action"].(string)
	if !ok {
		r.logMessage("Отсутствует поле action.")
		SendJSON(conn, "Missing action field")
		return
	}

	if action == "get_public_key" {
		r.logMessage("Получен запрос на публичный ключ. Отправка...")
		response := map[string]string{
			"e": r.publicKey["e"].String(),
			"n": r.publicKey["n"].String(),
		}
		if !SendJSON(conn, response) {
			r.logMessage("Ошибка отправки публичного ключа.")
		} else {
			r.logMessage("Публичный ключ отправлен.")
		}
		return
	}

	if action == "send_key" {
		r.logMessage("Получено сообщение с зашифрованным ключом.")
		encryptedData, ok := message["encrypted_data"].([]interface{})
		if !ok {
			r.logMessage("Неверный формат зашифрованных данных.")
			SendJSON(conn, "Invalid encrypted data format")
			return
		}

		senderPublicKeyData, ok := message["sender_public_key"].(map[string]interface{})
		if !ok {
			r.logMessage("Неверный формат публичного ключа отправителя.")
			SendJSON(conn, "Invalid sender public key format")
			return
		}

		nStr, ok := senderPublicKeyData["n"].(string)
		eStr, ok2 := senderPublicKeyData["e"].(string)
		if !ok || !ok2 {
			r.logMessage("Ошибка в формате публичного ключа отправителя.")
			SendJSON(conn, "Invalid sender public key data")
			return
		}

		senderN, success := new(big.Int).SetString(nStr, 10)
		senderE, success2 := new(big.Int).SetString(eStr, 10)
		if !success || !success2 {
			r.logMessage("Не удалось преобразовать публичный ключ отправителя.")
			SendJSON(conn, "Error parsing sender public key")
			return
		}
		senderPublicKey := map[string]*big.Int{"n": senderN, "e": senderE}

		encryptedBlocks := make([]*big.Int, len(encryptedData))
		for i, block := range encryptedData {
			blockStr, ok := block.(string)
			if !ok {
				r.logMessage("Неверный формат блока зашифрованных данных.")
				SendJSON(conn, "Invalid encrypted block format")
				return
			}
			blockInt, success := new(big.Int).SetString(blockStr, 10)
			if !success {
				r.logMessage("Не удалось преобразовать блок в big.Int.")
				SendJSON(conn, "Error parsing encrypted block")
				return
			}
			encryptedBlocks[i] = blockInt
		}

		decryptedData := cipher.DecryptRSA(encryptedBlocks, r.privateKey["d"], r.privateKey["n"])
		if decryptedData == "" {
			r.logMessage("Ошибка расшифровки данных.")
			SendJSON(conn, "Decryption failed")
			return
		}
		r.logMessage("Данные расшифрованы.")

		var decryptedMessage map[string]interface{}
		if err := json.Unmarshal([]byte(decryptedData), &decryptedMessage); err != nil {
			r.logMessage(fmt.Sprintf("Ошибка десериализации расшифрованных данных: %v", err))
			SendJSON(conn, "Error parsing decrypted data")
			return
		}

		keyHex, ok := decryptedMessage["key"].(string)
		if !ok {
			r.logMessage("Неверный формат ключа.")
			SendJSON(conn, "Invalid key format")
			return
		}
		timestamp, ok := decryptedMessage["timestamp"].(float64)
		if !ok {
			r.logMessage("Неверный формат временной метки.")
			SendJSON(conn, "Invalid timestamp format")
			return
		}
		signatureHex, ok := decryptedMessage["signature"].(string)
		if !ok {
			r.logMessage("Неверный формат подписи.")
			SendJSON(conn, "Invalid signature format")
			return
		}

		keyBytes, err := hex.DecodeString(keyHex)
		if err != nil {
			r.logMessage(fmt.Sprintf("Ошибка декодирования ключа: %v", err))
			SendJSON(conn, "Error decoding key")
			return
		}
		signature, err := hex.DecodeString(signatureHex)
		if err != nil {
			r.logMessage(fmt.Sprintf("Ошибка декодирования подписи: %v", err))
			SendJSON(conn, "Error decoding signature")
			return
		}

		r.logMessage(fmt.Sprintf("Получен ключ: %x", keyBytes))
		r.logMessage(fmt.Sprintf("Получена временная метка: %d", int64(timestamp)))
		r.logMessage(fmt.Sprintf("Получена подпись: %x", signature))

		messageToVerify := append(keyBytes, []byte(fmt.Sprintf("%d", int64(timestamp)))...)
		pubKeyBytes, err := json.Marshal(map[string]string{
			"e": r.publicKey["e"].String(),
			"n": r.publicKey["n"].String(),
		})
		if err != nil {
			r.logMessage(fmt.Sprintf("Ошибка сериализации публичного ключа: %v", err))
			SendJSON(conn, "Error serializing public key")
			return
		}
		messageToVerify = append(messageToVerify, pubKeyBytes...)
		hash := hashf.Sha512(messageToVerify)

		if verifySignature(hash[:], signature, senderPublicKey["e"], senderPublicKey["n"]) {
			r.logMessage("Проверка подписи успешна.")
			SendJSON(conn, "Key exchange successful")
		} else {
			r.logMessage("Проверка подписи не удалась.")
			SendJSON(conn, "Key exchange failed")
		}
	}
}

func main() {
	receiver := NewReceiver()
	receiver.Run()
}