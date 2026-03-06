package main

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

type UserData struct {
	RValue   int   `json:"r_value"`
	GCoeffs  []int `json:"g_coeffs"`
}

var (
	aCoeffsMatrix [][]int
	registeredUsers = make(map[string]UserData)
	assignedRValues = make(map[int]bool)
	dataLock        = &sync.Mutex{}
)

// матрица a_ij 
func initializeGlobalParameters() {
	rand.Seed(time.Now().UnixNano())
	aCoeffsMatrix = make([][]int, DEFAULT_DEGREE_M+1)
	for i := range aCoeffsMatrix {
		aCoeffsMatrix[i] = make([]int, DEFAULT_DEGREE_M+1)
	}
	for i := 0; i <= DEFAULT_DEGREE_M; i++ {
		for j := i; j <= DEFAULT_DEGREE_M; j++ {
			val := rand.Intn(DEFAULT_PRIME)
			aCoeffsMatrix[i][j] = val
			if i != j {
				aCoeffsMatrix[j][i] = val
			}
		}
	}
	fmt.Printf("[TA] Инициализированы глобальные параметры: Prime=%d, Degree m=%d\n", DEFAULT_PRIME, DEFAULT_DEGREE_M)
	fmt.Println("[TA] Сгенерирована секретная матрица коэффициентов f(x,y) (a_ij).")
}

// g(x) 
func calculateUserGPolynomialCoeffs(rUserPublicID int) []int {
	gUserCoeffs := make([]int, DEFAULT_DEGREE_M+1)
	for k := 0; k <= DEFAULT_DEGREE_M; k++ {
		coeffForYK := 0
		rPowI := 1
		for i := 0; i <= DEFAULT_DEGREE_M; i++ {
			term := (aCoeffsMatrix[i][k] * rPowI) % DEFAULT_PRIME
			coeffForYK = (coeffForYK + term) % DEFAULT_PRIME
			rPowI = (rPowI * rUserPublicID) % DEFAULT_PRIME
		}
		gUserCoeffs[k] = coeffForYK
	}
	return gUserCoeffs
}


func handleClientRequest(conn net.Conn) {
	clientID := conn.RemoteAddr().String()
	fmt.Printf("[TA] Принято соединение от %s\n", clientID)
	defer func() {
		fmt.Printf("[TA] Закрытие соединения с %s\n", clientID)
		conn.Close()
	}()

	requestData, err := receiveJSONMessage(conn)
	if err != nil {
		fmt.Printf("[TA] Ошибка получения данных от %s: %v\n", clientID, err)
		return
	}

	request, ok := requestData.(map[string]interface{})
	if !ok {
		sendJSONMessage(conn, map[string]interface{}{"status": "error", "message": "Неверный формат запроса"})
		return
	}

	command, _ := request["command"].(string)
	userID, _ := request["user_id"].(string)
	response := map[string]interface{}{"status": "error", "message": "Неверная команда или отсутствуют параметры"}

	if command == "REGISTER" {
		if userID == "" {
			response = map[string]interface{}{"status": "error", "message": "Требуется user_id для REGISTER"}
		} else {
			dataLock.Lock()
			if userData, exists := registeredUsers[userID]; exists {
				response = map[string]interface{}{
					"status":    "ok",
					"message":   "Пользователь уже зарегистрирован. Отправка существующих параметров.",
					"user_id":   userID,
					"g_coeffs":  userData.GCoeffs,
					"r_self":    userData.RValue,
					"prime":     DEFAULT_PRIME,
					"degree_m":  DEFAULT_DEGREE_M,
				}
				fmt.Printf("[TA] Пользователь %s повторно запросил регистрацию. Отправлены существующие параметры.\n", userID)
			} else {
				newRVal := rand.Intn(DEFAULT_PRIME-1) + 1
				for assignedRValues[newRVal] {
					newRVal = rand.Intn(DEFAULT_PRIME-1) + 1
				}
				assignedRValues[newRVal] = true
				gCoeffs := calculateUserGPolynomialCoeffs(newRVal)
				registeredUsers[userID] = UserData{RValue: newRVal, GCoeffs: gCoeffs}
				response = map[string]interface{}{
					"status":    "ok",
					"message":   "Пользователь успешно зарегистрирован.",
					"user_id":   userID,
					"g_coeffs":  gCoeffs,
					"r_self":    newRVal,
					"prime":     DEFAULT_PRIME,
					"degree_m":  DEFAULT_DEGREE_M,
				}
				fmt.Printf("[TA] Зарегистрирован пользователь %s с r_id=%d\n", userID, newRVal)
			}
			dataLock.Unlock()
		}
	} else if command == "GET_PEER_INFO" {
		peerID, _ := request["peer_id"].(string)
		if peerID == "" {
			response = map[string]interface{}{"status": "error", "message": "Требуется peer_id для GET_PEER_INFO"}
		} else {
			dataLock.Lock()
			if peerData, exists := registeredUsers[peerID]; exists {
				response = map[string]interface{}{
					"status":    "ok",
					"peer_id":   peerID,
					"r_peer":    peerData.RValue,
					"prime":     DEFAULT_PRIME,
					"degree_m":  DEFAULT_DEGREE_M,
				}
				fmt.Printf("[TA] Отправлена информация о партнере %s (r_id=%d) для %s (запрос от %s)\n", peerID, peerData.RValue, clientID, userID)
			} else {
				response = map[string]interface{}{"status": "error", "message": fmt.Sprintf("Партнер %s не зарегистрирован", peerID)}
				fmt.Printf("[TA] Запрос информации о незарегистрированном партнере %s от %s (запрос от %s)\n", peerID, clientID, userID)
			}
			dataLock.Unlock()
		}
	}

	if err := sendJSONMessage(conn, response); err != nil {
		fmt.Printf("[TA] Ошибка отправки ответа клиенту %s: %v\n", clientID, err)
	}
}

func startServer() {
	initializeGlobalParameters()
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", TA_HOST, TA_PORT))
	if err != nil {
		fmt.Printf("[TA] FATAL: Не удалось привязаться к %s:%d - %v\n", TA_HOST, TA_PORT, err)
		fmt.Println("[TA] Убедитесь, что порт не занят.")
		return
	}
	defer listener.Close()
	fmt.Printf("[TA] Сервер доверенного центра слушает на %s:%d\n", TA_HOST, TA_PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("[TA] Ошибка принятия соединения: %v\n", err)
			continue
		}
		go handleClientRequest(conn)
	}
}

func main() {
	startServer()
}