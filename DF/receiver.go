package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sample-app/mylib"
	"net"
	"crypto/rand"
	"os"
)

type Receiver struct {
	role string
}

func NewReceiver() *Receiver {
	return &Receiver{role: "Receiver B"}
}

func (r *Receiver) logMessage(message string) {
	fmt.Printf("[%s] %s\n", r.role, message)
}

func (r *Receiver) runReceiverClient() {
	
	paramsData, err := os.ReadFile(PARAMS_FILE)
	if err != nil {
		r.logMessage(fmt.Sprintf("Ошибка чтения файла %s: %v", PARAMS_FILE, err))
		return
	}
	var params map[string]string
	if err := json.Unmarshal(paramsData, &params); err != nil {
		r.logMessage(fmt.Sprintf("Ошибка десериализации параметров: %v", err))
		return
	}

	p, ok := new(big.Int).SetString(params["p"], 10)
	if !ok {
		r.logMessage("Ошибка преобразования p в big.Int")
		return
	}
	g, ok := new(big.Int).SetString(params["g"], 10)
	if !ok {
		r.logMessage("Ошибка преобразования g в big.Int")
		return
	}


	y, err := rand.Int(rand.Reader, new(big.Int).Sub(p, big.NewInt(2)))
	if err != nil {
		r.logMessage(fmt.Sprintf("Ошибка генерации y: %v", err))
		return
	}
	y.Add(y, big.NewInt(2))


	B := mylib.FastPowMod(g, y, p)

	
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", HOST, PORT))
	if err != nil {
		r.logMessage(fmt.Sprintf("Ошибка подключения: %v", err))
		return
	}
	defer conn.Close()
	r.logMessage(fmt.Sprintf("Подключено к %s:%d", HOST, PORT))

	
	if !SendJSON(conn, map[string]string{"B": B.String()}) {
		r.logMessage("Ошибка отправки B")
		return
	}


	data, err := ReceiveJSON(conn)
	if err != nil {
		r.logMessage(fmt.Sprintf("Ошибка получения данных: %v", err))
		return
	}
	message, ok := data.(map[string]interface{})
	if !ok {
		r.logMessage("Неверный формат сообщения")
		return
	}
	AStr, ok := message["A"].(string)
	if !ok {
		r.logMessage("Отсутствует поле A")
		return
	}
	A, ok := new(big.Int).SetString(AStr, 10)
	if !ok {
		r.logMessage("Ошибка преобразования A в big.Int")
		return
	}

	//k = A^y mod p
	sharedKey := mylib.FastPowMod(A, y, p)
	r.logMessage(fmt.Sprintf("Общий ключ: %s", sharedKey.String()))
}

func main() {
	receiver := NewReceiver()
	receiver.runReceiverClient()
}