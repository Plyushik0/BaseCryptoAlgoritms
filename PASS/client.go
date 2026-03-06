package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (

	userStateDir  = "user_states"
)


type UserState struct {
	UserID        string `json:"user_id"`
	SecretW       string `json:"secret_w_hex"`
	TotalN        int    `json:"total_n"`
	CurrentAttempt int    `json:"current_attempt"`
	HashFunc      string `json:"hash_func"`
}

func loadUserState(userID string) (*UserState, error) {
	stateFile := filepath.Join(userStateDir, fmt.Sprintf("%s.state", userID))
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка чтения состояния для %s: %v", userID, err)
	}

	var state UserState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("ошибка декодирования состояния для %s: %v", userID, err)
	}
	return &state, nil
}

func saveUserState(state *UserState) error {
	stateFile := filepath.Join(userStateDir, fmt.Sprintf("%s.state", state.UserID))
	data, err := json.MarshalIndent(state, "", "    ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации состояния для %s: %v", state.UserID, err)
	}
	if err := os.MkdirAll(userStateDir, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории %s: %v", userStateDir, err)
	}
	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("ошибка записи состояния для %s: %v", state.UserID, err)
	}
	return nil
}

func sendRequest(message string) (string, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverHost, serverPort))
	if err != nil {
		return "", fmt.Errorf("ошибка подключения: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte(message)); err != nil {
		return "", fmt.Errorf("ошибка отправки запроса: %v", err)
	}

	data := make([]byte, 1024)
	n, err := conn.Read(data)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа: %v", err)
	}
	return string(data[:n]), nil
}

func registerUser() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Введите User ID: ")
	userID, _ := reader.ReadString('\n')
	userID = strings.TrimSpace(userID)

	fmt.Print("Введите секретный пароль (w): ")
	secretW, _ := reader.ReadString('\n')
	secretW = strings.TrimSpace(secretW)

	fmt.Print("Введите общее количество попыток (n): ")
	totalNStr, _ := reader.ReadString('\n')
	totalNStr = strings.TrimSpace(totalNStr)
	totalN, err := parseInt(totalNStr)
	if err != nil || totalN <= 0 {
		fmt.Println("Ошибка: Неверное количество попыток")
		return
	}

	fmt.Println("Доступные хеш-функции: sha256, sha512, streebog512")
	fmt.Print("Выберите хеш-функцию: ")
	hashFuncName, _ := reader.ReadString('\n')
	hashFuncName = strings.TrimSpace(hashFuncName)
	if !isValidHashFunc(hashFuncName) {
		fmt.Println("Ошибка: Неподдерживаемая хеш-функция")
		return
	}

	fmt.Printf("Генерация цепочки хешей длиной %d...\n", totalN)
	secretWBytes := []byte(secretW)
	w0, err := IterateHash(secretWBytes, totalN, hashFuncName)
	if err != nil {
		fmt.Printf("Ошибка вычисления w0: %v\n", err)
		return
	}
	w0Hex := BytesToHex(w0)

	request := fmt.Sprintf("REGISTER|%s|%d|%s|%s", userID, totalN, w0Hex, hashFuncName)
	response, err := sendRequest(request)
	if err != nil {
		fmt.Printf("Ошибка отправки запроса: %v\n", err)
		return
	}
	fmt.Printf("Ответ сервера: %s\n", response)

	if response == "REGISTER_SUCCESS" {
		state := &UserState{
			UserID:        userID,
			SecretW:       BytesToHex(secretWBytes),
			TotalN:        totalN,
			CurrentAttempt: 1,
			HashFunc:      hashFuncName,
		}
		if err := saveUserState(state); err != nil {
			fmt.Printf("Ошибка сохранения состояния: %v\n", err)
			return
		}
		fmt.Printf("Регистрация успешна. Состояние сохранено в %s\n", filepath.Join(userStateDir, fmt.Sprintf("%s.state", userID)))
		fmt.Printf("У вас есть %d попыток\n", totalN)
	} else {
		fmt.Println("Регистрация не удалась")
	}
}

func loginUser() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Введите User ID: ")
	userID, _ := reader.ReadString('\n')
	userID = strings.TrimSpace(userID)

	state, err := loadUserState(userID)
	if err != nil || state == nil {
		fmt.Printf("Ошибка: Пользователь %s не зарегистрирован или файл состояния отсутствует\n", userID)
		return
	}

	if state.CurrentAttempt > state.TotalN {
		fmt.Printf("Вход не удался для %s: Максимальное количество попыток (%d) достигнуто\n", userID, state.TotalN)
		return
	}

	fmt.Printf("Пользователь: %s, Попытка: %d/%d, Хеш: %s\n", userID, state.CurrentAttempt, state.TotalN, state.HashFunc)

	secretWBytes, err := HexToBytes(state.SecretW)
	if err != nil {
		fmt.Printf("Ошибка декодирования секретного пароля: %v\n", err)
		return
	}

	iterations := state.TotalN - state.CurrentAttempt
	fmt.Printf("Вычисление пароля для попытки %d (%d итераций хеширования)...\n", state.CurrentAttempt, iterations)
	wi, err := IterateHash(secretWBytes, iterations, state.HashFunc)
	if err != nil {
		fmt.Printf("Ошибка вычисления пароля: %v\n", err)
		return
	}
	wiHex := BytesToHex(wi)

	request := fmt.Sprintf("AUTH|%s|%d|%s|%s", userID, state.CurrentAttempt, wiHex, state.HashFunc)
	response, err := sendRequest(request)
	if err != nil {
		fmt.Printf("Ошибка отправки запроса: %v\n", err)
		return
	}
	fmt.Printf("Ответ сервера: %s\n", response)

	if response == "AUTH_SUCCESS" {
		state.CurrentAttempt++
		if err := saveUserState(state); err != nil {
			fmt.Printf("Ошибка сохранения состояния: %v\n", err)
			return
		}
		fmt.Printf("Вход успешен. Осталось попыток: %d\n", state.TotalN-state.CurrentAttempt+1)
	} else {
		fmt.Println("Вход не удался")
	}
}

func isValidHashFunc(hashFunc string) bool {
	return hashFunc == "sha256" || hashFunc == "sha512" || hashFunc == "streebog512"
}

func mainClient() {
	for {
		fmt.Println("\n--- Клиент одноразовых паролей ---")
		fmt.Println("1. Регистрация")
		fmt.Println("2. Вход")
		fmt.Println("3. Выход")
		fmt.Print("Выберите действие: ")

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			registerUser()
		case "2":
			loginUser()
		case "3":
			return
		default:
			fmt.Println("Неверный выбор")
		}
	}
}