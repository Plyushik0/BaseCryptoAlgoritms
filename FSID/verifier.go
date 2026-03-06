package main

import (
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"time"

	"sample-app/mylib"

)

type Verifier struct {
	role string
	n    *big.Int
}

func NewVerifier() *Verifier {
	return &Verifier{role: "Verifier B"}
}

func (v *Verifier) logMessage(message string) {
	fmt.Printf("[%s] %s\n", v.role, message)
}

func (v *Verifier) getNFromTC() bool {
	v.logMessage(fmt.Sprintf("Подключение к доверенному центру на %s:%d для получения n...", TC_HOST, TC_PORT))
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", TC_HOST, TC_PORT))
	if err != nil {
		v.logMessage(fmt.Sprintf("Ошибка подключения: %v", err))
		return false
	}
	defer conn.Close()

	nInterface, err := ReceiveJSON(conn)
	if err != nil {
		v.logMessage(fmt.Sprintf("Ошибка получения n: %v", err))
		return false
	}


	nStr, ok := nInterface.(string)
	if !ok {
		v.logMessage(fmt.Sprintf("Полученное n не является строкой: %v", nInterface))
		return false
	}

	n, success := new(big.Int).SetString(nStr, 10)
	if !success {
		v.logMessage("Не удалось преобразовать n в big.Int")
		return false
	}
	v.n = n
	v.logMessage(fmt.Sprintf("Успешно получено n от доверенного центра: %s", nStr))
	return true
}

func (v *Verifier) Run() {
	if !v.getNFromTC() {
		v.logMessage("Не удалось получить n от доверенного центра. Завершение.")
		return
	}

	v.logMessage("Процесс Verifier B начат.")
	var conn net.Conn
	var err error
	retries := 5
	for i := 0; i < retries; i++ {
		v.logMessage(fmt.Sprintf("Подключение к Prover A на %s:%d... Попытка %d/%d", PROVER_HOST, PROVER_PORT, i+1, retries))
		conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", PROVER_HOST, PROVER_PORT), 5*time.Second)
		if err == nil {
			v.logMessage("Успешно подключено к Prover A.")
			break
		}
		v.logMessage(fmt.Sprintf("Ошибка подключения: %v. Повтор через 5 секунд...", err))
		time.Sleep(5 * time.Second)
	}
	if conn == nil {
		v.logMessage("Не удалось подключиться к Prover A после нескольких попыток. Завершение.")
		return
	}
	defer conn.Close()

	vInterface, err := ReceiveJSON(conn)
	if err != nil {
		v.logMessage(fmt.Sprintf("Ошибка получения v: %v", err))
		return
	}
	vStr, ok := vInterface.(string)
	if !ok {
		v.logMessage(fmt.Sprintf("Полученное v не является строкой: %v", vInterface))
		return
	}
	vVal, success := new(big.Int).SetString(vStr, 10)
	if !success {
		v.logMessage("Не удалось преобразовать v в big.Int")
		return
	}
	v.logMessage(fmt.Sprintf("Получено v от Prover A: %s", vStr))

	v.logMessage(fmt.Sprintf("Начало %d итераций протокола.", ITERATIONS))
	identificationSuccessful := true

	for i := 0; i < ITERATIONS; i++ {
		v.logMessage(fmt.Sprintf("--- Итерация %d/%d ---", i+1, ITERATIONS))

		xInterface, err := ReceiveJSON(conn)
		if err != nil {
			v.logMessage(fmt.Sprintf("Ошибка получения x: %v", err))
			identificationSuccessful = false
			break
		}
		xStr, ok := xInterface.(string)
		if !ok {
			v.logMessage(fmt.Sprintf("Полученное x не является строкой: %v", xInterface))
			identificationSuccessful = false
			break
		}
		x, success := new(big.Int).SetString(xStr, 10)
		if !success {
			v.logMessage("Не удалось преобразовать x в big.Int")
			identificationSuccessful = false
			break
		}
		v.logMessage(fmt.Sprintf("Получено x от Prover A: %s", xStr))

		c := rand.Intn(2)
		if !SendJSON(conn, float64(c)) {
			v.logMessage("Ошибка отправки вызова c для Prover A.")
			identificationSuccessful = false
			break
		}
		v.logMessage(fmt.Sprintf("Сгенерирован и отправлен вызов c: %d", c))

		yInterface, err := ReceiveJSON(conn)
		if err != nil {
			v.logMessage(fmt.Sprintf("Ошибка получения y: %v", err))
			identificationSuccessful = false
			break
		}
		yStr, ok := yInterface.(string)
		if !ok {
			v.logMessage(fmt.Sprintf("Полученное y не является строкой: %v", yInterface))
			identificationSuccessful = false
			break
		}
		y, success := new(big.Int).SetString(yStr, 10)
		if !success {
			v.logMessage("Не удалось преобразовать y в big.Int")
			identificationSuccessful = false
			break
		}
		v.logMessage(fmt.Sprintf("Получено y от Prover A: %s", yStr))

		vC := mylib.FastPowMod(vVal, big.NewInt(int64(c)), v.n)
		expectedYSquared := new(big.Int).Mul(x, vC).Mod(new(big.Int).Mul(x, vC), v.n)
		actualYSquared := mylib.FastPowMod(y, big.NewInt(2), v.n)

		v.logMessage("Проверка условия: y != 0 и y^2 (mod n) == (x * v^c) (mod n)")
		v.logMessage(fmt.Sprintf("  Проверка 1 (y != 0): %v", y.Cmp(big.NewInt(0)) != 0))
		v.logMessage(fmt.Sprintf("  Проверка 2 (y^2 == expected): %s == %s (%v)", actualYSquared.String(), expectedYSquared.String(), actualYSquared.Cmp(expectedYSquared) == 0))

		if y.Cmp(big.NewInt(0)) != 0 && actualYSquared.Cmp(expectedYSquared) == 0 {
			v.logMessage(fmt.Sprintf("Итерация %d пройдена.", i+1))
		} else {
			v.logMessage(fmt.Sprintf("Итерация %d не пройдена! Prover, вероятно, обманывает или произошла ошибка.", i+1))
			identificationSuccessful = false
			break
		}
		v.logMessage(fmt.Sprintf("--- Конец итерации %d/%d ---", i+1, ITERATIONS))
	}

	if identificationSuccessful {
		v.logMessage(fmt.Sprintf("Успешно завершены все %d итераций.", ITERATIONS))
		fmt.Println("\n### VERIFIER: ИДЕНТИФИКАЦИЯ УСПЕШНА ###\n")
	} else {
		v.logMessage("Идентификация не удалась.")
		fmt.Println("\n### VERIFIER: ИДЕНТИФИКАЦИЯ НЕ УДАЛАСЬ ###\n")
	}

	v.logMessage("Процесс Verifier B завершен.")
}

func main() {
	verifier := NewVerifier()
	verifier.Run()
}