package main

import (
	"encoding/hex"
	"fmt"
	"sample-app/hashf"
	"math/big"
	"sample-app/mylib"
	"net"
	"os"
	"encoding/json"
	"crypto/rand"
)

type UserB struct {
	role string
}

func NewUserB() *UserB {
	return &UserB{role: "User B"}
}

func (u *UserB) logMessage(message string) {
	fmt.Printf("[Пользователь B] %s\n", message)
}

func (u *UserB) runSpekeResponder(p *big.Int, sharedPasswordInt *big.Int, userID string) {
	u.logMessage(fmt.Sprintf("Начало обмена ключами SPEKE как ответчик для '%s'...", userID))

	passwordBytes := []byte(hex.EncodeToString(sharedPasswordInt.Bytes()))
	gHash := hashf.Sha512(passwordBytes)
	gHashInt := new(big.Int).SetBytes(gHash[:])
	g := new(big.Int).Mod(gHashInt, p)
	if g.Cmp(big.NewInt(2)) < 0 {
		u.logMessage(fmt.Sprintf("Вычисленный g (%s) тривиален. Используется g=2 для устойчивости SPEKE.", g.String()))
		g.SetInt64(2)
	}
	u.logMessage(fmt.Sprintf("SPEKE g = %s (получен из OTP, хеширован с %s)", g.String(), SPEKE_PASSWORD_HASH_FUNCTION))

	y, err := rand.Int(rand.Reader, new(big.Int).Sub(p, big.NewInt(2)))
	if err != nil {
		u.logMessage(fmt.Sprintf("Ошибка генерации y: %v", err))
		return
	}
	y.Add(y, big.NewInt(2))
	BSpeke := mylib.FastPowMod(g, y, p)
	u.logMessage(fmt.Sprintf("Сгенерирован приватный y_B, публичный B_speke = %s", BSpeke.String()))

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", HOST, SPEKE_PORT))
	if err != nil {
		u.logMessage(fmt.Sprintf("Ошибка привязки сокета: %v", err))
		return
	}
	defer listener.Close()
	u.logMessage(fmt.Sprintf("Служба SPEKE слушает на %s:%d для %s...", HOST, SPEKE_PORT, userID))

	conn, err := listener.Accept()
	if err != nil {
		u.logMessage(fmt.Sprintf("Ошибка принятия подключения: %v", err))
		return
	}
	defer conn.Close()
	u.logMessage(fmt.Sprintf("Подключение SPEKE от %s. Ожидание A_speke.", conn.RemoteAddr().String()))

	data, err := ReceiveJSON(conn)
	if err != nil {
		u.logMessage(fmt.Sprintf("Ошибка получения данных: %v", err))
		return
	}
	message, ok := data.(map[string]interface{})
	if !ok || message["A_speke"] == nil {
		u.logMessage("Ошибка SPEKE - A_speke получен некорректно")
		return
	}
	ASpekeStr, ok := message["A_speke"].(string)
	if !ok {
		u.logMessage("Неверный формат A_speke")
		return
	}
	ASpeke, ok := new(big.Int).SetString(ASpekeStr, 10)
	if !ok {
		u.logMessage("Ошибка преобразования A_speke в big.Int")
		return
	}
	u.logMessage(fmt.Sprintf("Получен A_speke = %s. Отправка B_speke.", ASpeke.String()))

	if !SendJSON(conn, map[string]string{"B_speke": BSpeke.String()}) {
		u.logMessage("Ошибка отправки B_speke")
		return
	}

	sharedKey := mylib.FastPowMod(ASpeke, y, p)
	u.logMessage(fmt.Sprintf("Общий ключ SPEKE K = %s", sharedKey.String()))
	u.logMessage("Обмен ключами SPEKE успешен.")
}

func (u *UserB) runOTPServiceForOneClient(host string, port int, stateDir string) (string, *big.Int, string, string) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		u.logMessage(fmt.Sprintf("Ошибка привязки сокета OTP: %v", err))
		return "FAILURE", nil, "", ""
	}
	defer listener.Close()
	u.logMessage(fmt.Sprintf("Служба OTP слушает на %s:%d для одного клиента...", host, port))

	conn, err := listener.Accept()
	if err != nil {
		u.logMessage(fmt.Sprintf("Ошибка принятия подключения OTP: %v", err))
		return "FAILURE", nil, "", ""
	}
	defer conn.Close()
	u.logMessage(fmt.Sprintf("Подключение OTP от %s", conn.RemoteAddr().String()))

	data, err := ReceiveJSON(conn)
	if err != nil {
		u.logMessage(fmt.Sprintf("Ошибка получения данных OTP: %v", err))
		SendJSON(conn, map[string]interface{}{"status": "FAILURE", "message": "Invalid request"})
		return "FAILURE", nil, "", ""
	}
	message, ok := data.(map[string]interface{})
	if !ok || message["command"] == nil {
		u.logMessage("Неверный формат запроса OTP")
		SendJSON(conn, map[string]interface{}{"status": "FAILURE", "message": "Invalid request"})
		return "FAILURE", nil, "", ""
	}

	command, _ := message["command"].(string)
	userID, _ := message["user_id"].(string)
	hashFunc, _ := message["hash_func"].(string)
	var response map[string]interface{}
	var wiForSpeke *big.Int

	if command == "REGISTER" {
		response = handleOTPRegisterB(message, stateDir)
		SendJSON(conn, response)
		return response["status"].(string), nil, userID, hashFunc
	} else if command == "AUTH" {
		response, wiForSpeke = handleOTPAuthB(message, stateDir)
		SendJSON(conn, response)
		if response["status"] == "AUTH_SUCCESS" {
			return "AUTH_SUCCESS", wiForSpeke, userID, hashFunc
		}
		return response["status"].(string), nil, userID, hashFunc
	}
	u.logMessage("Неизвестная команда OTP")
	SendJSON(conn, map[string]interface{}{"status": "FAILURE", "message": "Unknown command"})
	return "FAILURE", nil, "", ""
}

func (u *UserB) mainUserB() {
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

	u.logMessage("--- Служба Пользователя B ---")
	otpStatus, wiInt, userID, otpHashUsed := u.runOTPServiceForOneClient(HOST, PORT, USER_B_OTP_SERVER_STATE_DIR)

	if otpStatus == "AUTH_SUCCESS" && wiInt != nil && userID != "" {
		u.logMessage(fmt.Sprintf("Аутентификация OTP успешна для пользователя '%s'.", userID))
		u.logMessage(fmt.Sprintf("Пароль OTP для использования в SPEKE (wi_int): %s", wiInt.String()))
		u.logMessage(fmt.Sprintf("Использовалась хеш-функция OTP: %s", otpHashUsed))
		u.runSpekeResponder(p, wiInt, userID)
	} else if otpStatus == "REGISTER_SUCCESS" {
		u.logMessage(fmt.Sprintf("Регистрация OTP завершена для пользователя '%s'.", userID))
		u.logMessage("Для установки ключа Пользователь A должен войти с OTP в следующий раз.")
	} else {
		u.logMessage(fmt.Sprintf("Процесс OTP завершился со статусом: %s. Нельзя продолжить SPEKE.", otpStatus))
	}
	u.logMessage("Работа службы завершена.")
}

func main() {
	userB := NewUserB()
	userB.mainUserB()
}