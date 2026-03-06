package main

import (
	"flag"
	"fmt"
	"net"
	"strings"
)

func requestFromTA(payload map[string]interface{}) (map[string]interface{}, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", TA_HOST, TA_PORT))
	if err != nil {
		return nil, fmt.Errorf("[UserClient] CRITICAL: Подключение к TA (%s:%d) отклонено: %v. Запущен ли сервер TA?", TA_HOST, TA_PORT, err)
	}
	defer conn.Close()

	if err := sendJSONMessage(conn, payload); err != nil {
		return nil, err
	}

	response, err := receiveJSONMessage(conn)
	if err != nil {
		return nil, err
	}

	responseMap, ok := response.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("[UserClient] Неверный формат ответа от TA")
	}
	return responseMap, nil
}

func main() {
	userIDPtr := flag.String("id", "", "Идентификатор пользователя (например, A, B). Регистр не учитывается, будет преобразован в верхний.")
	peerIDPtr := flag.String("peer", "", "Идентификатор партнера для установки ключа. Регистр не учитывается.")
	flag.Parse()

	userID := strings.ToUpper(*userIDPtr)
	peerID := strings.ToUpper(*peerIDPtr)

	if userID == "" || peerID == "" {
		fmt.Println("[UserClient] Ошибка: Необходимо указать --id и --peer")
		flag.Usage()
		return
	}
	if userID == peerID {
		fmt.Println("[UserClient] Ошибка: Идентификатор пользователя и партнера не могут совпадать.")
		return
	}

	fmt.Printf("--- Клиент пользователя: '%s' (Попытка установить ключ с '%s') ---\n", userID, peerID)

	fmt.Printf("\n[Шаг 1] Регистрация '%s' в доверенном центре...\n", userID)
	registerPayload := map[string]interface{}{"command": "REGISTER", "user_id": userID}
	regResponse, err := requestFromTA(registerPayload)
	if err != nil {
		fmt.Printf("[UserClient] Ошибка регистрации '%s': %v\n", userID, err)
		return
	}
	if regResponse["status"] != "ok" {
		fmt.Printf("[UserClient] Регистрация не удалась для '%s'.\n", userID)
		fmt.Printf("   Ответ TA: %v\n", regResponse["message"])
		return
	}

	gCoeffs, _ := regResponse["g_coeffs"].([]interface{})
	myRValue, _ := regResponse["r_self"].(float64)
	prime, _ := regResponse["prime"].(float64)
	if gCoeffs == nil || myRValue == 0 || prime == 0 {
		fmt.Printf("[UserClient] Неполные данные регистрации для '%s'. Ответ: %v\n", userID, regResponse)
		return
	}

	gCoeffsInt := make([]int, len(gCoeffs))
	for i, coeff := range gCoeffs {
		gCoeffsInt[i] = int(coeff.(float64))
	}

	fmt.Printf("[UserClient] Успешная регистрация '%s' (или получение существующих параметров).\n", userID)
	fmt.Printf("   Мой публичный ID (r_%s): %d\n", userID, int(myRValue))
	fmt.Printf("   Коэффициенты моего секретного полинома g_%s(x): %v\n", userID, gCoeffsInt)
	fmt.Printf("   Простое число системы (p): %d\n", int(prime))

	fmt.Printf("\n[Шаг 2] Запрос публичной информации о партнере '%s' из доверенного центра...\n", peerID)
	getPeerPayload := map[string]interface{}{"command": "GET_PEER_INFO", "user_id": userID, "peer_id": peerID}
	peerResponse, err := requestFromTA(getPeerPayload)
	if err != nil {
		fmt.Printf("[UserClient] Ошибка получения информации о партнере '%s': %v\n", peerID, err)
		return
	}
	if peerResponse["status"] != "ok" {
		fmt.Printf("[UserClient] Не удалось получить информацию о партнере '%s'.\n", peerID)
		fmt.Printf("   Ответ TA: %v\n", peerResponse["message"])
		fmt.Printf("   Подсказка: Убедитесь, что партнер '%s' зарегистрирован в TA.\n", peerID)
		return
	}

	rPeer, _ := peerResponse["r_peer"].(float64)
	if rPeer == 0 {
		fmt.Printf("[UserClient] Неполные данные о партнере '%s'. Ответ: %v\n", peerID, peerResponse)
		return
	}

	fmt.Printf("[UserClient] Успешно получена публичная информация о партнере '%s'.\n", peerID)
	fmt.Printf("   Публичный ID партнера '%s' (r_%s): %d\n", peerID, peerID, int(rPeer))

	fmt.Printf("\n[Шаг 3] Вычисление общего ключа K_(%s,%s) = g_%s(r_%s)...\n", userID, peerID, userID, peerID)
	sharedKey := polyEval(gCoeffsInt, int(rPeer), int(prime))
	fmt.Printf("   Вычисление: g_%s(%d) mod %d\n", userID, int(rPeer), int(prime))
	fmt.Printf("   ---> ОБЩИЙ КЛЮЧ для '%s' с '%s': %d <---\n", userID, peerID, sharedKey)
}