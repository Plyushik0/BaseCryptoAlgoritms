package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
)

// восстанавливает секрет с помощью интерполяции Лагранжа
func reconstructSecret(shares []map[string]*big.Int, prime *big.Int) (*big.Int, error) {
	if len(shares) == 0 {
		return nil, fmt.Errorf("подмножество долей не может быть пустым")
	}

	secretSum := big.NewInt(0)
	for i, shareI := range shares {
		xI, yI := shareI["x"], shareI["y"]
		numerator, denominator := big.NewInt(1), big.NewInt(1)

		for j, shareJ := range shares {
			if i == j {
				continue
			}
			xJ := shareJ["x"]
			numerator.Mul(numerator, xJ).Mod(numerator, prime)
			// diff = x_j - x_i mod prime
			diff := new(big.Int).Sub(xJ, xI) // diff = x_j - x_i
			diff.Mod(diff, prime)            // Приводим к [0, prime)
			if diff.Cmp(big.NewInt(0)) == 0 {
				return nil, fmt.Errorf("дублирующиеся x-координаты в долях")
			}
			denominator.Mul(denominator, diff).Mod(denominator, prime)
		}

		if denominator.Cmp(big.NewInt(0)) == 0 {
			return nil, fmt.Errorf("знаменатель равен нулю при интерполяции Лагранжа")
		}

		lagrangeBasis, err := modularInverse(denominator, prime)
		if err != nil {
			return nil, err
		}
		lagrangeBasis.Mul(lagrangeBasis, numerator).Mod(lagrangeBasis, prime)
		term := new(big.Int).Mul(yI, lagrangeBasis)
		term.Mod(term, prime)
		secretSum.Add(secretSum, term).Mod(secretSum, prime)
	}
	return secretSum, nil
}

func getShareFromString(jsonStr string) map[string]*big.Int {
	var response map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &response)
	result := make(map[string]*big.Int)
	for key, value := range response["share_data"].(map[string]interface{}) {
		result[key], _ = new(big.Int).SetString(value.(string), 10)
	}
	return result
}

func getShareInfoFromServer(conn net.Conn) (int64, int64, *big.Int, error) {
	request := map[string]interface{}{"action": "get_share_info"}
	if err := sendJSON(conn, request); err != nil {
		return 0, 0, nil, err
	}

	response, err := receiveJSON(conn)
	if err != nil {
		return 0, 0, nil, err
	}
	if response["status"] != "success" {
		return 0, 0, nil, fmt.Errorf("не удалось получить информацию о долях: %v", response["message"])
	}

	n, _ := response["n"].(float64)
	t, _ := response["t"].(float64)
	primeStr, _ := response["prime"].(string)
	prime, ok := new(big.Int).SetString(primeStr, 10)
	if !ok {
		return 0, 0, nil, fmt.Errorf("неверный формат prime: %v", primeStr)
	}
	return int64(n), int64(t), prime, nil
}

func getShareFromServer(conn net.Conn, shareID int64) (map[string]*big.Int, error) {
	request := map[string]interface{}{"action": "get_share", "share_id": shareID}
	if err := sendJSON(conn, request); err != nil {
		return nil, err
	}

	response, err := receiveJSON(conn)
	if err != nil {
		return nil, err
	}
	if response["status"] != "success" {
		return nil, fmt.Errorf("не удалось получить долю %d: %v", shareID, response["message"])
	}

	shareData, _ := response["share_data"].(map[string]interface{})
	share := make(map[string]*big.Int)
	for key, value := range shareData {
		val, ok := new(big.Int).SetString(value.(string), 10)
		if !ok {
			return nil, fmt.Errorf("неверный формат значения %s: %v", key, value)
		}
		share[key] = val
	}
	return share, nil
}

func main() {
	fmt.Println("Запуск клиента-комбайнера...")


	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", HOST, PORT))
	if err != nil {
		fmt.Printf("Подключение отклонено. Запущен ли сервер дилера на %s:%d? %v\n", HOST, PORT, err)
		return
	}
	defer conn.Close()

	n, t, prime, err := getShareInfoFromServer(conn)
	if err != nil {
		fmt.Printf("Не удалось получить информацию о долях: %v\n", err)
		return
	}
	fmt.Printf("\nПараметры схемы от сервера: N=%d, Порог T=%d, Prime=%v\n", n, t, prime)

	fmt.Printf("Введите не менее %d ID долей для восстановления, разделенных запятыми (например, 1,3,5): ", t)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	shareIDsStr := scanner.Text()
	shareIDs := make([]int64, 0)
	for _, sidStr := range strings.Split(shareIDsStr, ",") {
		sid, err := strconv.ParseInt(strings.TrimSpace(sidStr), 10, 64)
		if err != nil {
			fmt.Println("Неверный формат ID долей. Введите числа, разделенные запятыми.")
			return
		}
		shareIDs = append(shareIDs, sid)
	}

	if len(shareIDs) < int(t) {
		fmt.Printf("Необходимо не менее %d долей для восстановления секрета. Вы указали %d.\n", t, len(shareIDs))
		return
	}

	uniqueIDs := make(map[int64]bool)
	for _, sid := range shareIDs {
		if sid < 1 || sid > n {
			fmt.Printf("Неверный ID доли: %d. ID должны быть от 1 до %d.\n", sid, n)
			return
		}
		if uniqueIDs[sid] {
			fmt.Printf("Дублирующийся ID доли: %d.\n", sid)
			return
		}
		uniqueIDs[sid] = true
	}

	fmt.Printf("\nЗапрос долей для ID: %v...\n", shareIDs)
	collectedShares := make([]map[string]*big.Int, 0)
	for _, shareID := range shareIDs {
		share, err := getShareFromServer(conn, shareID)
		if err != nil {
			fmt.Printf("Не удалось получить долю %d: %v\n", shareID, err)
			continue
		}
		fmt.Printf("Успешно получена доля %d: %v\n", shareID, share)
		collectedShares = append(collectedShares, share)
	}

	if len(collectedShares) < int(t) {
		fmt.Printf("Недостаточно долей (%d собрано, требуется %d). Восстановление невозможно.\n", len(collectedShares), t)
		return
	}

	fmt.Printf("\nПопытка восстановить секрет, используя %d долей: %v\n", len(collectedShares), collectedShares)
	secret, err := reconstructSecret(collectedShares, prime)
	if err != nil {
		fmt.Printf("Ошибка восстановления секрета: %v\n", err)
		return
	}
	fmt.Printf("\nВосстановленный секрет: %v\n", secret)
}