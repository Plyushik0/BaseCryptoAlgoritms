package main

import (
	"encoding/json"
	"fmt"
	"sample-app/mylib"
	"os"
)

func generateAndSaveSpekeParams() {
	fmt.Printf("Генерация %d-битного простого числа p для SPEKE...\n", PRIME_BITS_SPEKE)
	p := mylib.GeneratePrimeBitK(PRIME_BITS_SPEKE, 40) 
	params := map[string]string{
		"p": p.String(),
	}
	data, err := json.Marshal(params)
	if err != nil {
		fmt.Printf("Ошибка сериализации параметров: %v\n", err)
		return
	}
	if err := os.WriteFile(SPEKE_PARAMS_FILE, data, 0644); err != nil {
		fmt.Printf("Ошибка записи в файл %s: %v\n", SPEKE_PARAMS_FILE, err)
		return
	}
	fmt.Printf("Параметры SPEKE (p) сохранены в %s\n", SPEKE_PARAMS_FILE)
	fmt.Printf("p = %s\n", p.String())
}

func clearFile() {
	if err := os.WriteFile(SPEKE_PARAMS_FILE, []byte{}, 0644); err != nil {
		fmt.Printf("Ошибка очистки файла %s: %v\n", SPEKE_PARAMS_FILE, err)
	}
}

func main() {
	if err := os.MkdirAll(USER_A_OTP_STATE_DIR, 0755); err != nil {
		fmt.Printf("Ошибка создания директории %s: %v\n", USER_A_OTP_STATE_DIR, err)
		return
	}
	if err := os.MkdirAll(USER_B_OTP_SERVER_STATE_DIR, 0755); err != nil {
		fmt.Printf("Ошибка создания директории %s: %v\n", USER_B_OTP_SERVER_STATE_DIR, err)
		return
	}
	clearFile()
	generateAndSaveSpekeParams()
}