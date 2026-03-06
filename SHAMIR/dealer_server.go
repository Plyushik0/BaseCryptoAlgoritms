package main

import (
	"bufio"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"encoding/json"
)

type Share struct {
	ID int64     `json:"id"`
	X  *big.Int `json:"x"`
	Y  *big.Int `json:"y"`
}

type SharesData struct {
	Shares []Share   `json:"shares"`
	N      int64     `json:"n"`
	T      int64     `json:"t"`
	Prime  *big.Int `json:"prime"`
}

var (
	globalSharesData SharesData
	dataLock         = &sync.Mutex{}
)

// генерирует n долей для секрета с порогом t
func generateShares(secret *big.Int, n, t int64, prime *big.Int) ([]Share, []*big.Int, error) {
	if !(1 < t && t <= n) {
		return nil, nil, fmt.Errorf("порог t=%d должен удовлетворять 1 < t <= n=%d", t, n)
	}
	if secret.Cmp(big.NewInt(0)) < 0 || secret.Cmp(prime) >= 0 {
		return nil, nil, fmt.Errorf("секрет %v должен быть в диапазоне [0, %v)", secret, prime)
	}
	if big.NewInt(n).Cmp(prime) >= 0 {
		return nil, nil, fmt.Errorf("число долей n=%d должно быть меньше prime=%v", n, prime)
	}

	degree := int(t - 1)
	coeffs, err := generateRandomCoefficients(degree, prime, secret)
	if err != nil {
		return nil, nil, err
	}

	shares := make([]Share, n)
	for i := int64(1); i <= n; i++ {
		xVal := big.NewInt(i)
		yVal := evaluatePolynomial(coeffs, xVal, prime)
		shares[i-1] = Share{ID: i, X: xVal, Y: yVal}
	}
	return shares, coeffs, nil
}


func saveSharesToFiles(shares []Share, n, t int64, prime *big.Int) error {
	if err := os.MkdirAll(SHARES_DIR, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию %s: %v", SHARES_DIR, err)
	}

	metadata := map[string]interface{}{
		"n":     n,
		"t":     t,
		"prime": prime.String(),
	}
	metadataBytes, _ := json.MarshalIndent(metadata, "", "  ")
	if err := os.WriteFile(METADATA_FILE, metadataBytes, 0644); err != nil {
		return fmt.Errorf("не удалось сохранить метаданные в %s: %v", METADATA_FILE, err)
	}
	fmt.Printf("Метаданные сохранены в %s\n", METADATA_FILE)

	for _, share := range shares {
		shareFile := filepath.Join(SHARES_DIR, fmt.Sprintf("share_%d.json", share.ID))
		shareContent := map[string]string{"x": share.X.String(), "y": share.Y.String()}
		shareBytes, _ := json.MarshalIndent(shareContent, "", "  ")
		if err := os.WriteFile(shareFile, shareBytes, 0644); err != nil {
			return fmt.Errorf("не удалось сохранить долю %d в %s: %v", share.ID, shareFile, err)
		}
		fmt.Printf("Доля %d сохранена в %s\n", share.ID, shareFile)
	}
	return nil
}


func handleClient(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	fmt.Printf("Подключен клиент %s\n", addr)
	defer func() {
		fmt.Printf("Закрытие соединения с %s\n", addr)
		conn.Close()
	}()

	for {
		request, err := receiveJSON(conn)
		if err != nil {
			fmt.Printf("Ошибка получения запроса от %s: %v\n", addr, err)
			return
		}
		fmt.Printf("Получен запрос от %s: %v\n", addr, request)

		action, _ := request["action"].(string)
		response := map[string]interface{}{"status": "error", "message": "Неверный запрос"}

		dataLock.Lock()
		if action == "get_share_info" {
			if globalSharesData.N == 0 {
				response = map[string]interface{}{"status": "error", "message": "Доли еще не сгенерированы"}
			} else {
				response = map[string]interface{}{
					"status": "success",
					"n":      globalSharesData.N,
					"t":      globalSharesData.T,
					"prime":  globalSharesData.Prime.String(),
				}
			}
		} else if action == "get_share" {
			shareIDFloat, _ := request["share_id"].(float64)
			shareID := int64(shareIDFloat)
			if globalSharesData.N == 0 {
				response = map[string]interface{}{"status": "error", "message": "Доли еще не сгенерированы"}
			} else if shareID < 1 || shareID > globalSharesData.N {
				response = map[string]interface{}{"status": "error", "message": fmt.Sprintf("Неверный ID доли: %d. Должен быть от 1 до %d", shareID, globalSharesData.N)}
			} else {
				for _, share := range globalSharesData.Shares {
					if share.ID == shareID {
						response = map[string]interface{}{
							"status":     "success",
							"share_id":   shareID,
							"share_data": map[string]string{"x": share.X.String(), "y": share.Y.String()},
						}
						break
					}
				}
				if response["status"] != "success" {
					response = map[string]interface{}{"status": "error", "message": fmt.Sprintf("Доля с ID %d не найдена", shareID)}
				}
			}
		} else if action == "ping" {
			response = map[string]interface{}{"status": "success", "message": "pong"}
		}
		dataLock.Unlock()

		if err := sendJSON(conn, response); err != nil {
			fmt.Printf("Ошибка отправки ответа клиенту %s: %v\n", addr, err)
			return
		}
		fmt.Printf("Отправлен ответ клиенту %s: %v\n", addr, response)
	}
}

func main() {
	fmt.Printf("Введите секрет (целое число < %v): ", PRIME)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	secret, ok := new(big.Int).SetString(scanner.Text(), 10)
	if !ok || secret.Cmp(big.NewInt(0)) < 0 || secret.Cmp(PRIME) >= 0 {
		fmt.Printf("Секрет должен быть целым числом от 0 до %v. Выход.\n", new(big.Int).Sub(PRIME, big.NewInt(1)))
		return
	}

	fmt.Printf("Введите общее количество долей (n, по умолчанию %d): ", DEFAULT_N)
	scanner.Scan()
	nStr := scanner.Text()
	n := DEFAULT_N
	if nStr != "" {
		n64, err := strconv.ParseInt(nStr, 10, 64)
		if err != nil {
			fmt.Printf("Неверное значение n. Используется значение по умолчанию: %d\n", DEFAULT_N)
			n = DEFAULT_N
		} else {
			n = n64
		}
	}

	fmt.Printf("Введите порог (t, по умолчанию %d): ", DEFAULT_T)
	scanner.Scan()
	tStr := scanner.Text()
	t := DEFAULT_T
	if tStr != "" {
		t64, err := strconv.ParseInt(tStr, 10, 64)
		if err != nil {
			fmt.Printf("Неверное значение t. Используется значение по упмолчанию: %d\n", DEFAULT_T)
			t = DEFAULT_T
		} else {
			t = t64
		}
	}

	if !(1 < t && t <= n) {
		fmt.Printf("Неверные параметры: должно быть 1 < t (%d) <= n (%d). Выход.\n", t, n)
		return
	}

	fmt.Printf("\nГенерация %d долей для секрета %v с порогом %d, используя prime=%v...\n", n, secret, t, PRIME)
	shares, _, err := generateShares(secret, n, t, PRIME)
	if err != nil {
		fmt.Printf("Ошибка генерации долей: %v\n", err)
		return
	}
	dataLock.Lock()
	globalSharesData = SharesData{Shares: shares, N: n, T: t, Prime: new(big.Int).Set(PRIME)}
	dataLock.Unlock()
	fmt.Printf("Успешно сгенерировано %d долей.\n", len(shares))

	if err := saveSharesToFiles(shares, n, t, PRIME); err != nil {
		fmt.Printf("Ошибка сохранения долей: %v\n", err)
		return
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", HOST, PORT))
	if err != nil {
		fmt.Printf("FATAL: Не удалось привязаться к %s:%d - %v\n", HOST, PORT, err)
		return
	}
	defer listener.Close()
	fmt.Printf("\nСервер дилера слушает на %s:%d\n", HOST, PORT)
	fmt.Println("Готов раздавать доли клиентам.")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Ошибка принятия соединения: %v\n", err)
			continue
		}
		go handleClient(conn)
	}
}