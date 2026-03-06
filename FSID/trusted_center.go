package main

import (
	"fmt"
	"math/big"
	"net"
	"sample-app/mylib"

)

type TrustedCenter struct {
	role string
}

func NewTrustedCenter() *TrustedCenter {
	return &TrustedCenter{role: "Trusted Center T"}
}

func (tc *TrustedCenter) logMessage(message string) {
	fmt.Printf("[%s] %s\n", tc.role, message)
}

func (tc *TrustedCenter) Run() {
	tc.logMessage("Запуск доверенного центра...")


	tc.logMessage(fmt.Sprintf("Генерация простых чисел p и q (%d бит каждое)...", BIT_LENGTH))
	p := mylib.GeneratePrimeBitK(BIT_LENGTH, 10)
	q := mylib.GeneratePrimeBitK(BIT_LENGTH, 10)
	for p.Cmp(q) == 0 {
		tc.logMessage("p и q совпадают, регенерация q...")
		q = mylib.GeneratePrimeBitK(BIT_LENGTH, 10)
	}

	n := new(big.Int).Mul(p, q)
	tc.logMessage(fmt.Sprintf("Сгенерированы p и q. Вычислено n: %s", n.String()))

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", TC_HOST, TC_PORT))
	if err != nil {
		tc.logMessage(fmt.Sprintf("Ошибка привязки сокета: %v", err))
		return
	}
	defer listener.Close()
	tc.logMessage(fmt.Sprintf("Слушает на %s:%d", TC_HOST, TC_PORT))

	
	for i := 0; i < 2; i++ {
		conn, err := listener.Accept()
		if err != nil {
			tc.logMessage(fmt.Sprintf("Ошибка принятия подключения: %v", err))
			continue
		}
		role := "Prover A"
		if i == 1 {
			role = "Verifier B"
		}
		tc.logMessage(fmt.Sprintf("Принято подключение от %s на %s", role, conn.RemoteAddr().String()))

		
		nStr := n.String()
		if !SendJSON(conn, nStr) {
			tc.logMessage(fmt.Sprintf("Ошибка отправки n для %s", role))
		} else {
			tc.logMessage(fmt.Sprintf("Отправлено n для %s: %s", role, nStr))
		}
		conn.Close()
		tc.logMessage(fmt.Sprintf("Закрыто соединение с %s", role))
	}

	tc.logMessage("Доверенный центр завершил отправку n обеим сторонам.")
}

func main() {
	tc := NewTrustedCenter()
	tc.Run()
}