package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"os"
	"sample-app/hashf"
	"strconv"
)

type UserAState struct {
	UserID         string   `json:"user_id"`
	SecretWInt     *big.Int `json:"-"`
	SecretWIntHex  string   `json:"secret_w_int_hex"`
	TotalN         int      `json:"total_n"`
	CurrentAttempt int      `json:"current_attempt"`
	HashFunc       string   `json:"hash_func"`
}

type UserBState struct {
	UserID              string   `json:"user_id"`
	ExpectedAttempt     int      `json:"expected_attempt"`
	ExpectedPasswordInt *big.Int `json:"-"`
	ExpectedPasswordHex string   `json:"expected_password_hex"`
	HashFunc            string   `json:"hash_func"`
	TotalN              int      `json:"total_n"`
}

func loadUserAState(userID, stateDir string) (*UserAState, error) {
	stateFile := fmt.Sprintf("%s/%s.a_state", stateDir, userID)
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла состояния %s: %v", stateFile, err)
	}
	var state UserAState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("ошибка десериализации состояния %s: %v", userID, err)
	}
	secretWInt, ok := new(big.Int).SetString(state.SecretWIntHex, 16)
	if !ok {
		return nil, fmt.Errorf("ошибка преобразования secret_w_int_hex для %s", userID)
	}
	state.SecretWInt = secretWInt
	return &state, nil
}

func saveUserAState(state *UserAState, stateDir string) error {
	stateFile := fmt.Sprintf("%s/%s.a_state", stateDir, state.UserID)
	stateCopy := *state
	stateCopy.SecretWIntHex = hex.EncodeToString(state.SecretWInt.Bytes())
	data, err := json.MarshalIndent(stateCopy, "", "    ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации состояния %s: %v", state.UserID, err)
	}
	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("ошибка записи состояния %s: %v", stateFile, err)
	}
	return nil
}

func loadUserBState(userID, stateDir string) (*UserBState, error) {
	stateFile := fmt.Sprintf("%s/%s.b_server_state", stateDir, userID)
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла состояния %s: %v", stateFile, err)
	}
	var state UserBState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("ошибка десериализации состояния %s: %v", userID, err)
	}
	expectedPasswordInt, ok := new(big.Int).SetString(state.ExpectedPasswordHex, 16)
	if !ok {
		return nil, fmt.Errorf("ошибка преобразования expected_password_hex для %s", userID)
	}
	state.ExpectedPasswordInt = expectedPasswordInt
	return &state, nil
}

func saveUserBState(state *UserBState, stateDir string) error {
	stateFile := fmt.Sprintf("%s/%s.b_server_state", stateDir, state.UserID)
	stateCopy := *state
	stateCopy.ExpectedPasswordHex = hex.EncodeToString(state.ExpectedPasswordInt.Bytes())
	data, err := json.MarshalIndent(stateCopy, "", "    ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации состояния %s: %v", state.UserID, err)
	}
	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("ошибка записи состояния %s: %v", stateFile, err)
	}
	return nil
}

func iterateHash(input *big.Int, iterations int, hashFunc string) *big.Int {
	result := new(big.Int).Set(input)
	for i := 0; i < iterations; i++ {
		inputBytes := hex.EncodeToString(result.Bytes())
		var hash []byte
		if hashFunc == "sha256" {
			hashArr := hashf.Sha256([]byte(inputBytes))
			hash = hashArr[:]
		} else if hashFunc == "sha512" {
			hashArr := hashf.Sha512([]byte(inputBytes))
			hash = hashArr[:]
		} else {
			panic("unsupported hash function")
		}
		result.SetBytes(hash)
	}
	return result
}

func registerA(userID, secretWStr string, totalN int, hashFunc, stateDir, serverHost string, serverPort int) bool {
	fmt.Printf("Пользователь A: Регистрация '%s' у Пользователя B...\n", userID)
	secretWBytes := []byte(secretWStr)
	initialWInt := new(big.Int).SetBytes(secretWBytes)

	fmt.Printf("Пользователь A: Генерация цепочки хешей длиной %d для OTP с использованием %s...\n", totalN, hashFunc)
	w0Int := iterateHash(initialWInt, totalN, hashFunc)
	w0Hex := hex.EncodeToString(w0Int.Bytes())

	request := map[string]interface{}{
		"command":   "REGISTER",
		"user_id":   userID,
		"total_n":   fmt.Sprintf("%d", totalN),
		"w0_hex":    w0Hex,
		"hash_func": hashFunc,
	}
	response, err := sendOTPRequest(request, serverHost, serverPort)
	if err != nil {
		fmt.Printf("Пользователь A (Регистрация OTP): Ошибка отправки запроса: %v\n", err)
		return false
	}
	fmt.Printf("Пользователь A (Регистрация OTP): Пользователь B ответил: %v\n", response)

	responseMap, ok := response.(map[string]interface{})
	if !ok {
		fmt.Println("Пользователь A: Неверный формат ответа от Пользователя B")
		return false
	}
	if responseMap["status"] == "REGISTER_SUCCESS" {
		userState := &UserAState{
			UserID:         userID,
			SecretWInt:     initialWInt,
			TotalN:         totalN,
			CurrentAttempt: 1,
			HashFunc:       hashFunc,
		}
		if err := saveUserAState(userState, stateDir); err != nil {
			fmt.Printf("Пользователь A: Ошибка сохранения состояния: %v\n", err)
			return false
		}
		fmt.Println("Пользователь A: Регистрация OTP успешна. Состояние сохранено.")
		return true
	}
	fmt.Printf("Пользователь A: Регистрация OTP не удалась. Причина: %v\n", responseMap["message"])
	return false
}

func loginA(userID, stateDir, serverHost string, serverPort int) (*big.Int, string, error) {
	fmt.Printf("Пользователь A: Попытка входа OTP для '%s' у Пользователя B...\n", userID)
	userState, err := loadUserAState(userID, stateDir)
	if err != nil {
		fmt.Printf("Пользователь A: Ошибка - Пользователь '%s' не зарегистрирован локально или файл состояния отсутствует: %v\n", userID, err)
		return nil, "", err
	}

	totalN := userState.TotalN
	currentAttempt := userState.CurrentAttempt
	hashFunc := userState.HashFunc
	initialWInt := userState.SecretWInt

	if currentAttempt > totalN {
		fmt.Printf("Пользователь A: Вход не удался для '%s': Достигнуто максимальное количество попыток OTP (%d). Пожалуйста, зарегистрируйтесь заново.\n", userID, totalN)
		return nil, "", fmt.Errorf("max attempts reached")
	}

	fmt.Printf("Пользователь A: Попытка OTP: %d/%d, Хеш: %s\n", currentAttempt, totalN, hashFunc)
	iterations := totalN - currentAttempt
	fmt.Printf("Пользователь A: Вычисление пароля OTP для попытки %d (%d итераций хеширования)...\n", currentAttempt, iterations)

	wiInt := iterateHash(initialWInt, iterations, hashFunc)
	wiHex := hex.EncodeToString(wiInt.Bytes())

	request := map[string]interface{}{
		"command":     "AUTH",
		"user_id":     userID,
		"attempt_num": fmt.Sprintf("%d", currentAttempt),
		"wi_hex":      wiHex,
		"hash_func":   hashFunc,
	}
	response, err := sendOTPRequest(request, serverHost, serverPort)
	if err != nil {
		fmt.Printf("Пользователь A (Аутентификация OTP): Ошибка отправки запроса: %v\n", err)
		return nil, "", err
	}
	fmt.Printf("Пользователь A (Аутентификация OTP): Пользователь B ответил: %v\n", response)

	responseMap, ok := response.(map[string]interface{})
	if !ok {
		fmt.Println("Пользователь A: Неверный формат ответа от Пользователя B")
		return nil, "", fmt.Errorf("invalid response format")
	}
	if responseMap["status"] == "AUTH_SUCCESS" {
		userState.CurrentAttempt++
		if err := saveUserAState(userState, stateDir); err != nil {
			fmt.Printf("Пользователь A: Ошибка сохранения состояния: %v\n", err)
			return nil, "", err
		}
		fmt.Printf("Пользователь A: Вход OTP успешен. Осталось попыток: %d\n", totalN-userState.CurrentAttempt+1)
		return wiInt, hashFunc, nil
	}
	fmt.Printf("Пользователь A: Вход OTP не удался. Причина: %v\n", responseMap["message"])
	return nil, "", fmt.Errorf("auth failed: %v", responseMap["message"])
}

func handleOTPRegisterB(data map[string]interface{}, stateDir string) map[string]interface{} {
	userID, _ := data["user_id"].(string)
	totalNStr, _ := data["total_n"].(string)
	w0Hex, _ := data["w0_hex"].(string)
	hashFunc, _ := data["hash_func"].(string)
	fmt.Printf("Пользователь B: Попытка регистрации OTP для пользователя: %s\n", userID)

	if _, err := loadUserBState(userID, stateDir); err == nil {
		return map[string]interface{}{"status": "REGISTER_FAILURE", "message": "Пользователь уже зарегистрирован"}
	}

	totalN, err := strconv.Atoi(totalNStr)
	if err != nil || totalN <= 0 {
		return map[string]interface{}{"status": "REGISTER_FAILURE", "message": fmt.Sprintf("Недопустимое количество попыток: %v", err)}
	}
	w0Int, ok := new(big.Int).SetString(w0Hex, 16)
	if !ok {
		return map[string]interface{}{"status": "REGISTER_FAILURE", "message": "Недопустимый формат w0_hex"}
	}

	if hashFunc != "sha256" && hashFunc != "sha512" {
		return map[string]interface{}{"status": "REGISTER_FAILURE", "message": "Неподдерживаемая хеш-функция"}
	}

	serverState := &UserBState{
		UserID:              userID,
		ExpectedAttempt:     1,
		ExpectedPasswordInt: w0Int,
		HashFunc:            hashFunc,
		TotalN:              totalN,
	}
	if err := saveUserBState(serverState, stateDir); err != nil {
		return map[string]interface{}{"status": "REGISTER_FAILURE", "message": fmt.Sprintf("Ошибка сохранения состояния: %v", err)}
	}
	fmt.Printf("Пользователь B: Регистрация OTP для %s успешна.\n", userID)
	return map[string]interface{}{"status": "REGISTER_SUCCESS"}
}

func handleOTPAuthB(data map[string]interface{}, stateDir string) (map[string]interface{}, *big.Int) {
	userID, _ := data["user_id"].(string)
	attemptNumStr, _ := data["attempt_num"].(string)
	wiHex, _ := data["wi_hex"].(string)
	hashFunc, _ := data["hash_func"].(string)
	fmt.Printf("Пользователь B: Попытка аутентификации OTP для пользователя: %s\n", userID)

	serverState, err := loadUserBState(userID, stateDir)
	if err != nil {
		return map[string]interface{}{"status": "AUTH_FAILURE", "message": "Пользователь не зарегистрирован или ошибка состояния"}, nil
	}

	attemptNum, err := strconv.Atoi(attemptNumStr)
	if err != nil {
		return map[string]interface{}{"status": "AUTH_FAILURE", "message": fmt.Sprintf("Недопустимый номер попытки: %v", err)}, nil
	}
	wiInt, ok := new(big.Int).SetString(wiHex, 16)
	if !ok {
		return map[string]interface{}{"status": "AUTH_FAILURE", "message": "Недопустимый формат wi_hex"}, nil
	}

	if serverState.HashFunc != hashFunc {
		return map[string]interface{}{"status": "AUTH_FAILURE", "message": fmt.Sprintf("Несоответствие хеш-функции. Сервер ожидает %s, клиент отправил %s", serverState.HashFunc, hashFunc)}, nil
	}

	if attemptNum > serverState.TotalN {
		return map[string]interface{}{"status": "AUTH_FAILURE", "message": "Превышено максимальное количество попыток для этой регистрации"}, nil
	}

	if attemptNum != serverState.ExpectedAttempt {
		msg := fmt.Sprintf("Ожидалась попытка %d, получена %d.", serverState.ExpectedAttempt, attemptNum)
		if attemptNum < serverState.ExpectedAttempt {
			msg += " Возможна атака повторного воспроизведения или ошибка синхронизации."
		} else {
			msg += " Ошибка синхронизации."
		}
		return map[string]interface{}{"status": "AUTH_FAILURE", "message": msg}, nil
	}

	wiBytes := hex.EncodeToString(wiInt.Bytes())
	var hashedWi []byte
	if hashFunc == "sha256" {
		hashedWiArr := hashf.Sha256([]byte(wiBytes))
		hashedWi = hashedWiArr[:]
	} else if hashFunc == "sha512" {
		hashedWiArr := hashf.Sha512([]byte(wiBytes))
		hashedWi = hashedWiArr[:]
	} else {
		return map[string]interface{}{"status": "AUTH_FAILURE", "message": "Неподдерживаемая хеш-функция"}, nil
	}
	hashedWiInt := new(big.Int).SetBytes(hashedWi)

	if hashedWiInt.Cmp(serverState.ExpectedPasswordInt) == 0 {
		fmt.Printf("Пользователь B: Аутентификация OTP успешна для %s (попытка %d).\n", userID, attemptNum)
		serverState.ExpectedAttempt++
		serverState.ExpectedPasswordInt = wiInt
		if err := saveUserBState(serverState, stateDir); err != nil {
			return map[string]interface{}{"status": "AUTH_FAILURE", "message": fmt.Sprintf("Ошибка сохранения состояния: %v", err)}, nil
		}
		return map[string]interface{}{"status": "AUTH_SUCCESS"}, wiInt
	}
	fmt.Printf("Пользователь B: Аутентификация OTP не удалась для %s: Несоответствие пароля.\n", userID)
	return map[string]interface{}{"status": "AUTH_FAILURE", "message": "Несоответствие пароля"}, nil
}

func sendOTPRequest(message map[string]interface{}, host string, port int) (interface{}, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения: %v", err)
	}
	defer conn.Close()
	if !SendJSON(conn, message) {
		return nil, fmt.Errorf("ошибка отправки запроса")
	}
	return ReceiveJSON(conn)
}
