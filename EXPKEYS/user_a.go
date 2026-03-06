package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"sample-app/hashf"
	"math/big"
	"sample-app/mylib"
	"os"
	"strconv"
	"strings"
	"net"
	"encoding/json"
	"crypto/rand"
)

type UserA struct {
	role   string
	userID string
}

func NewUserA() *UserA {
	return &UserA{role: "User A"}
}

func (u *UserA) logMessage(message string) {
	fmt.Printf("[Пользователь A] %s\n", message)
}

func (u *UserA) runSpekeInitiator(p *big.Int, sharedPasswordInt *big.Int, userID string) {
	u.userID = userID
	u.logMessage(fmt.Sprintf("Начало обмена ключами SPEKE для '%s'...", userID))

	passwordBytes := []byte(hex.EncodeToString(sharedPasswordInt.Bytes()))
	gHash := hashf.Sha512(passwordBytes)
	gHashInt := new(big.Int).SetBytes(gHash[:])
	g := new(big.Int).Mod(gHashInt, p)
	if g.Cmp(big.NewInt(2)) < 0 {
		u.logMessage(fmt.Sprintf("Вычисленный g (%s) тривиален. Используется g=2 для устойчивости SPEKE.", g.String()))
		g.SetInt64(2)
	}
	u.logMessage(fmt.Sprintf("SPEKE g = %s (получен из OTP, хеширован с %s)", g.String(), SPEKE_PASSWORD_HASH_FUNCTION))

	x, err := rand.Int(rand.Reader, new(big.Int).Sub(p, big.NewInt(2)))
	if err != nil {
		u.logMessage(fmt.Sprintf("Ошибка генерации x: %v", err))
		return
	}
	x.Add(x, big.NewInt(2))
	ASpeke := mylib.FastPowMod(g, x, p)
	u.logMessage(fmt.Sprintf("Сгенерирован приватный x_A, публичный A_speke = %s", ASpeke.String()))

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", HOST, SPEKE_PORT))
	if err != nil {
		u.logMessage(fmt.Sprintf("Ошибка SPEKE - Подключение отклонено Пользователем B на порту %d: %v", SPEKE_PORT, err))
		return
	}
	defer conn.Close()
	u.logMessage(fmt.Sprintf("Подключено к порту SPEKE Пользователя B %d. Отправка A_speke.", SPEKE_PORT))

	if !SendJSON(conn, map[string]string{"A_speke": ASpeke.String()}) {
		u.logMessage("Ошибка отправки A_speke")
		return
	}

	u.logMessage("Ожидание B_speke от Пользователя B...")
	data, err := ReceiveJSON(conn)
	if err != nil {
		u.logMessage(fmt.Sprintf("Ошибка получения данных: %v", err))
		return
	}
	message, ok := data.(map[string]interface{})
	if !ok || message["B_speke"] == nil {
		u.logMessage("Ошибка SPEKE - B_speke получен некорректно")
		return
	}
	BSpekeStr, ok := message["B_speke"].(string)
	if !ok {
		u.logMessage("Неверный формат B_speke")
		return
	}
	BSpeke, ok := new(big.Int).SetString(BSpekeStr, 10)
	if !ok {
		u.logMessage("Ошибка преобразования B_speke в big.Int")
		return
	}
	u.logMessage(fmt.Sprintf("Получен B_speke = %s", BSpeke.String()))

	sharedKey := mylib.FastPowMod(BSpeke, x, p)
	u.logMessage(fmt.Sprintf("Общий ключ SPEKE K = %s", sharedKey.String()))
	u.logMessage("Обмен ключами SPEKE успешен.")
}

func (u *UserA) mainUserA() {
	paramsData, err := os.ReadFile(SPEKE_PARAMS_FILE)
	if err != nil {
		u.logMessage(fmt.Sprintf("Ошибка чтения файла %s: %v", SPEKE_PARAMS_FILE, err))
		return
	}
	var params map[string]string
	if err := json.Unmarshal(paramsData, &params); err != nil {
		u.logMessage(fmt.Sprintf("Ошибка десериализации параметров: %v", err))
		return
	}
	p, ok := new(big.Int).SetString(params["p"], 10)
	if !ok {
		u.logMessage("Ошибка преобразования p в big.Int")
		return
	}
	u.logMessage(fmt.Sprintf("Загружено простое число SPEKE p (длина %d бит).", p.BitLen()))

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n--- Меню Пользователя A ---")
		fmt.Println("1. Зарегистрироваться для OTP у Пользователя B")
		fmt.Println("2. Войти с OTP и установить ключ SPEKE с Пользователем B")
		fmt.Println("3. Выход")
		fmt.Print("Введите выбор: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		if choice == "1" {
			fmt.Print("Введите идентификатор пользователя для OTP: ")
			userID, _ := reader.ReadString('\n')
			userID = strings.TrimSpace(userID)
			fmt.Print("Введите ваш главный секретный пароль (w) для OTP: ")
			secretWStr, _ := reader.ReadString('\n')
			secretWStr = strings.TrimSpace(secretWStr)
			fmt.Print("Введите общее количество попыток OTP (n): ")
			totalNStr, _ := reader.ReadString('\n')
			totalNStr = strings.TrimSpace(totalNStr)
			totalN, err := strconv.Atoi(totalNStr)
			if err != nil || totalN <= 0 {
				u.logMessage(fmt.Sprintf("Недопустимый ввод: %v", err))
				continue
			}
			fmt.Printf("Используется хеш-функция OTP: %s\n", OTP_HASH_FUNCTION)
			registerA(userID, secretWStr, totalN, OTP_HASH_FUNCTION, USER_A_OTP_STATE_DIR, HOST, PORT)
		} else if choice == "2" {
			fmt.Print("Введите идентификатор пользователя для входа OTP: ")
			userID, _ := reader.ReadString('\n')
			userID = strings.TrimSpace(userID)
			wiInt, _, err := loginA(userID, USER_A_OTP_STATE_DIR, HOST, PORT)
			if err != nil {
				u.logMessage("Вход OTP не удался. Нельзя продолжить SPEKE.")
				continue
			}
			u.logMessage(fmt.Sprintf("Вход OTP успешен для %s. Пароль для SPEKE (wi_int): %s", userID, wiInt.String()))
			u.runSpekeInitiator(p, wiInt, userID)
		} else if choice == "3" {
			u.logMessage("Выход.")
			break
		} else {
			u.logMessage("Недопустимый выбор. Пожалуйста, попробуйте снова.")
		}
	}
}

func main() {
	userA := NewUserA()
	userA.mainUserA()
}