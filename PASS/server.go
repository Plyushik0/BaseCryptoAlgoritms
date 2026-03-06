package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	serverHost     = "127.0.0.1"
	serverPort     = 12345
	serverStateDir = "server_states"
)


type ServerState struct {
	UserID            string `json:"user_id"`
	ExpectedAttempt   int    `json:"expected_attempt"`
	ExpectedPassword  string `json:"expected_password_hex"`
	HashFunc          string `json:"hash_func"`
}

func loadServerState(userID string) (*ServerState, error) {
	stateFile := filepath.Join(serverStateDir, fmt.Sprintf("%s.serverstate", userID))
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка чтения состояния для %s: %v", userID, err)
	}

	var state ServerState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("ошибка декодирования состояния для %s: %v", userID, err)
	}
	return &state, nil
}

func saveServerState(state *ServerState) error {
	stateFile := filepath.Join(serverStateDir, fmt.Sprintf("%s.serverstate", state.UserID))
	data, err := json.MarshalIndent(state, "", "    ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации состояния для %s: %v", state.UserID, err)
	}
	if err := os.MkdirAll(serverStateDir, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории %s: %v", serverStateDir, err)
	}
	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("ошибка записи состояния для %s: %v", state.UserID, err)
	}
	return nil
}

func handleRegister(parts []string) string {
	if len(parts) != 5 {
		return "REGISTER_FAILURE|Неверный формат запроса"
	}
	_, userID, totalNStr, w0Hex, hashFuncName := parts[0], parts[1], parts[2], parts[3], parts[4]
	fmt.Printf("Попытка регистрации пользователя: %s\n", userID)

	if state, _ := loadServerState(userID); state != nil {
		fmt.Printf("Пользователь %s уже зарегистрирован\n", userID)
		return "REGISTER_FAILURE|Пользователь уже зарегистрирован"
	}

	totalN, err := parseInt(totalNStr)
	if err != nil || totalN <= 0 {
		return "REGISTER_FAILURE|Неверное количество попыток"
	}


	_, err = HashMessage([]byte{0}, hashFuncName)
	if err != nil {
		return fmt.Sprintf("REGISTER_FAILURE|Неподдерживаемая хеш-функция: %s", hashFuncName)
	}

	state := &ServerState{
		UserID:           userID,
		ExpectedAttempt:  1,
		ExpectedPassword: w0Hex,
		HashFunc:         hashFuncName,
	}
	if err := saveServerState(state); err != nil {
		fmt.Printf("Ошибка сохранения состояния для %s: %v\n", userID, err)
		return fmt.Sprintf("REGISTER_FAILURE|Ошибка сервера: %v", err)
	}

	fmt.Printf("Пользователь %s успешно зарегистрирован с %d попытками, используя %s\n", userID, totalN, hashFuncName)
	return "REGISTER_SUCCESS"
}

func handleAuth(parts []string) string {
	if len(parts) != 5 {
		return "AUTH_FAILURE|Неверный формат запроса"
	}
	_, userID, attemptNumStr, wiHex, hashFuncName := parts[0], parts[1], parts[2], parts[3], parts[4]
	fmt.Printf("Попытка аутентификации пользователя: %s\n", userID)

	state, err := loadServerState(userID)
	if err != nil || state == nil {
		fmt.Printf("Аутентификация не удалась для %s: Пользователь не зарегистрирован или файл состояния поврежден\n", userID)
		return "AUTH_FAILURE|Пользователь не зарегистрирован или файл состояния поврежден"
	}

	attemptNum, err := parseInt(attemptNumStr)
	if err != nil {
		fmt.Printf("Ошибка обработки запроса для %s: %v\n", userID, err)
		return fmt.Sprintf("AUTH_FAILURE|Неверный формат данных: %v", err)
	}

	if state.HashFunc != hashFuncName {
		fmt.Printf("Аутентификация не удалась для %s: Несоответствие хеш-функции\n", userID)
		return "AUTH_FAILURE|Несоответствие хеш-функции"
	}

	if attemptNum != state.ExpectedAttempt {
		fmt.Printf("Аутентификация не удалась для %s: Ожидалась попытка %d, получена %d\n", userID, state.ExpectedAttempt, attemptNum)
		if attemptNum < state.ExpectedAttempt {
			return "AUTH_FAILURE|Номер попытки слишком низкий (возможная атака повторного воспроизведения)"
		}
		return "AUTH_FAILURE|Номер попытки слишком высокий (ошибка синхронизации)"
	}

	wiBytes, err := HexToBytes(wiHex)
	if err != nil {
		fmt.Printf("Ошибка обработки wi для %s: %v\n", userID, err)
		return fmt.Sprintf("AUTH_FAILURE|Неверный формат пароля: %v", err)
	}

	hashedWi, err := HashMessage(wiBytes, hashFuncName)
	if err != nil {
		fmt.Printf("Ошибка хеширования для %s: %v\n", userID, err)
		return fmt.Sprintf("AUTH_FAILURE|Ошибка хеширования: %v", err)
	}

	expectedPassword, err := HexToBytes(state.ExpectedPassword)
	if err != nil {
		fmt.Printf("Ошибка обработки ожидаемого пароля для %s: %v\n", userID, err)
		return fmt.Sprintf("AUTH_FAILURE|Ошибка формата ожидаемого пароля: %v", err)
	}

	if string(hashedWi) != string(expectedPassword) {
		fmt.Printf("Аутентификация не удалась для %s: Пароль не совпадает\n", userID)
		return "AUTH_FAILURE|Пароль не совпадает"
	}


	fmt.Printf("Аутентификация успешна для %s на попытке %d\n", userID, attemptNum)
	state.ExpectedAttempt++
	state.ExpectedPassword = wiHex
	if err := saveServerState(state); err != nil {
		fmt.Printf("Ошибка сохранения состояния для %s: %v\n", userID, err)
		return fmt.Sprintf("AUTH_SUCCESS|Но ошибка сохранения состояния: %v", err)
	}

	return "AUTH_SUCCESS"
}

func startServer() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", serverHost, serverPort))
	if err != nil {
		fmt.Printf("Ошибка запуска сервера: %v\n", err)
		return
	}
	defer listener.Close()
	fmt.Printf("Центр аутентификации слушает на %s:%d\n", serverHost, serverPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Ошибка принятия соединения: %v\n", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	data := make([]byte, 1024)
	n, err := conn.Read(data)
	if err != nil {
		fmt.Printf("Ошибка чтения данных: %v\n", err)
		conn.Write([]byte("FAILURE|Ошибка сервера: " + err.Error()))
		return
	}

	message := string(data[:n])
	parts := strings.Split(message, "|")
	var response string

	if len(parts) == 0 {
		response = "FAILURE|Неверный формат сообщения"
	} else if parts[0] == "REGISTER" {
		response = handleRegister(parts)
	} else if parts[0] == "AUTH" {
		response = handleAuth(parts)
	} else {
		response = "FAILURE|Неизвестная команда"
	}

	if _, err := conn.Write([]byte(response)); err != nil {
		fmt.Printf("Ошибка отправки ответа: %v\n", err)
	}
	fmt.Printf("Отправлен ответ: %s\n", response)
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}