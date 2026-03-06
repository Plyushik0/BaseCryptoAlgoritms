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



type Sender struct {
	role string
}

func NewSender() *Sender {
	return &Sender{role: "Sender A"}
}

func (s *Sender) logMessage(message string) {
	fmt.Printf("[%s] %s\n", s.role, message)
}

func (s *Sender) runSenderServer() {
	
	paramsData, err := os.ReadFile(PARAMS_FILE)
	if err != nil {
		s.logMessage(fmt.Sprintf("Ошибка чтения файла %s: %v", PARAMS_FILE, err))
		return
	}
	var params map[string]string
	if err := json.Unmarshal(paramsData, &params); err != nil {
		s.logMessage(fmt.Sprintf("Ошибка десериализации параметров: %v", err))
		return
	}

	p, ok := new(big.Int).SetString(params["p"], 10)
	if !ok {
		s.logMessage("Ошибка преобразования p в big.Int")
		return
	}
	g, ok := new(big.Int).SetString(params["g"], 10)
	if !ok {
		s.logMessage("Ошибка преобразования g в big.Int")
		return
	}

	
	x, err := rand.Int(rand.Reader, new(big.Int).Sub(p, big.NewInt(2)))
	if err != nil {
		s.logMessage(fmt.Sprintf("Ошибка генерации x: %v", err))
		return
	}
	x.Add(x, big.NewInt(2)) 

	//  A = g^x mod p
	A := mylib.FastPowMod(g, x, p)


	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", HOST, PORT))
	if err != nil {
		s.logMessage(fmt.Sprintf("Ошибка привязки сокета: %v", err))
		return
	}
	defer listener.Close()
	s.logMessage(fmt.Sprintf("Слушает на %s:%d", HOST, PORT))

	conn, err := listener.Accept()
	if err != nil {
		s.logMessage(fmt.Sprintf("Ошибка принятия подключения: %v", err))
		return
	}
	defer conn.Close()
	s.logMessage(fmt.Sprintf("Принято подключение от %s", conn.RemoteAddr().String()))

	
	data, err := ReceiveJSON(conn)
	if err != nil {
		s.logMessage(fmt.Sprintf("Ошибка получения данных: %v", err))
		return
	}
	message, ok := data.(map[string]interface{})
	if !ok {
		s.logMessage("Неверный формат сообщения")
		return
	}
	BStr, ok := message["B"].(string)
	if !ok {
		s.logMessage("Отсутствует поле B")
		return
	}
	B, ok := new(big.Int).SetString(BStr, 10)
	if !ok {
		s.logMessage("Ошибка преобразования B в big.Int")
		return
	}

	
	if !SendJSON(conn, map[string]string{"A": A.String()}) {
		s.logMessage("Ошибка отправки A")
		return
	}

	// k = B^x mod p
	sharedKey := mylib.FastPowMod(B, x, p)
	s.logMessage(fmt.Sprintf("Общий ключ: %s", sharedKey.String()))
}

func main() {
	sender := NewSender()
	sender.runSenderServer()
}