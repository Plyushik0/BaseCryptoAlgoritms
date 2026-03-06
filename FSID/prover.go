package main

import (
	"fmt"
	"math/big"
	"crypto/rand"
	"net"
	"sample-app/mylib"

)

type Prover struct {
	role string
	n    *big.Int
	s    *big.Int
	v    *big.Int
}

func NewProver() *Prover {
	return &Prover{role: "Prover A"}
}

func (p *Prover) logMessage(message string) {
	fmt.Printf("[%s] %s\n", p.role, message)
}

func (p *Prover) getNFromTC() bool {
	p.logMessage(fmt.Sprintf("Подключение к доверенному центру на %s:%d для получения n...", TC_HOST, TC_PORT))
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", TC_HOST, TC_PORT))
	if err != nil {
		p.logMessage(fmt.Sprintf("Ошибка подключения: %v", err))
		return false
	}
	defer conn.Close()

	nInterface, err := ReceiveJSON(conn)
	if err != nil {
		p.logMessage(fmt.Sprintf("Ошибка получения n: %v", err))
		return false
	}

	
	nStr, ok := nInterface.(string)
	if !ok {
		p.logMessage(fmt.Sprintf("Полученное n не является строкой: %v", nInterface))
		return false
	}

	n, success := new(big.Int).SetString(nStr, 10)
	if !success {
		p.logMessage("Не удалось преобразовать n в big.Int")
		return false
	}
	p.n = n
	p.logMessage(fmt.Sprintf("Успешно получено n от доверенного центра: %s", nStr))
	return true
}

func (p *Prover) generateSecret() bool {
	p.logMessage("Генерация секрета s (взаимно простого с n)...")
	one := big.NewInt(1)
	nMinusOne := new(big.Int).Sub(p.n, one)
	for {
		s, err := rand.Int(rand.Reader, nMinusOne)
		if err != nil {
			p.logMessage(fmt.Sprintf("Ошибка генерации s: %v", err))
			return false
		}
		if s.Cmp(one) <= 0 {
			continue
		}
		gcd, _, _ := mylib.ExtendedGCD(s, p.n)
		if gcd.Cmp(one) == 0 {
			p.s = s
			p.v = mylib.FastPowMod(s, big.NewInt(2), p.n)
			p.logMessage(fmt.Sprintf("Сгенерирован секрет s и вычислено v = s^2 (mod n): %s", p.v.String()))
			return true
		}
	}
}

func (p *Prover) Run() {
	if !p.getNFromTC() {
		p.logMessage("Не удалось получить n от доверенного центра. Завершение.")
		return
	}

	if !p.generateSecret() {
		p.logMessage("Не удалось сгенерировать секрет s. Завершение.")
		return
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", PROVER_HOST, PROVER_PORT))
	if err != nil {
		p.logMessage(fmt.Sprintf("Ошибка привязки сокета: %v", err))
		return
	}
	defer listener.Close()
	p.logMessage(fmt.Sprintf("Слушает для Verifier B на %s:%d", PROVER_HOST, PROVER_PORT))

	conn, err := listener.Accept()
	if err != nil {
		p.logMessage(fmt.Sprintf("Ошибка принятия подключения от Verifier B: %v", err))
		return
	}
	defer conn.Close()
	p.logMessage(fmt.Sprintf("Принято соединение от Verifier B на %s", conn.RemoteAddr().String()))

	if !SendJSON(conn, p.v.String()) {
		p.logMessage("Ошибка отправки v для Verifier B.")
		return
	}
	p.logMessage(fmt.Sprintf("Отправлено v для Verifier B: %s", p.v.String()))

	p.logMessage(fmt.Sprintf("Начало %d итераций протокола.", ITERATIONS))
	identificationSuccessful := true

	for i := 0; i < ITERATIONS; i++ {
		p.logMessage(fmt.Sprintf("--- Итерация %d/%d ---", i+1, ITERATIONS))

		zLimit := new(big.Int).Sub(p.n, big.NewInt(2))
		z, err := rand.Int(rand.Reader, zLimit)
		if err != nil {
			p.logMessage(fmt.Sprintf("Ошибка генерации случайного z: %v", err))
			identificationSuccessful = false
			break
		}
		z.Add(z, big.NewInt(2)) 
		p.logMessage("Сгенерировано случайное z.")
		x := mylib.FastPowMod(z, big.NewInt(2), p.n)
		p.logMessage(fmt.Sprintf("Вычислено x = z^2 (mod n): %s", x.String()))

		if !SendJSON(conn, x.String()) {
			p.logMessage("Ошибка отправки x для Verifier B.")
			identificationSuccessful = false
			break
		}
		p.logMessage(fmt.Sprintf("Отправлено x для Verifier B: %s", x.String()))

		cInterface, err := ReceiveJSON(conn)
		if err != nil {
			p.logMessage(fmt.Sprintf("Ошибка получения вызова c: %v", err))
			identificationSuccessful = false
			break
		}
		cFloat, ok := cInterface.(float64)
		if !ok {
			p.logMessage(fmt.Sprintf("Полученный c не является числом: %v", cInterface))
			identificationSuccessful = false
			break
		}
		c := int64(cFloat)
		if c != 0 && c != 1 {
			p.logMessage(fmt.Sprintf("Ошибка: Получен неверный вызов c: %d. Ожидается 0 или 1.", c))
			identificationSuccessful = false
			break
		}
		p.logMessage(fmt.Sprintf("Получен вызов c (%d) от Verifier B.", c))

		var y *big.Int
		if c == 0 {
			y = z
			p.logMessage("c равно 0, y = z.")
		} else {
			y = new(big.Int).Mul(z, p.s)
			y.Mod(y, p.n)
			p.logMessage("c равно 1, y = (z * s) mod n.")
		}

		if !SendJSON(conn, y.String()) {
			p.logMessage("Ошибка отправки y для Verifier B.")
			identificationSuccessful = false
			break
		}
		p.logMessage(fmt.Sprintf("Отправлено y для Verifier B: %s", y.String()))
		p.logMessage(fmt.Sprintf("--- Конец итерации %d/%d ---", i+1, ITERATIONS))
	}

	if identificationSuccessful {
		p.logMessage("Процесс Prover успешно завершил свою часть.")
		fmt.Println("\n### PROVER: ИДЕНТИФИКАЦИЯ УСПЕШНА ###\n")
	} else {
		p.logMessage("Процесс Prover завершился с ошибкой во время итераций.")
		fmt.Println("\n### PROVER: ИДЕНТИФИКАЦИЯ НЕ УДАЛАСЬ ###\n")
	}

	p.logMessage("Процесс Prover A завершен.")
}

func main() {
	prover := NewProver()
	prover.Run()
}