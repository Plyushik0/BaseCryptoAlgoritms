package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
    "time"
    "sample-app/cipher"
    "math/big"
)

func main() {
    fmt.Println("--- Симуляция протокола идентификации с RSA ---")


    if _, err := os.Stat(PUB_KEY_A); os.IsNotExist(err) {
        _, _, _, _ = cipher.GenerateRSAKeys(1024, PUB_KEY_A, PRIV_KEY_A)
        println("Ключи для A сгенерированы и сохранены.")
    }
    if _, err := os.Stat(PUB_KEY_B); os.IsNotExist(err) {
        _, _, _, _ = cipher.GenerateRSAKeys(1024, PUB_KEY_B, PRIV_KEY_B)
        println("Ключи для B сгенерированы и сохранены.")
    }

    fmt.Println("\nВведите сообщение для протокола:")
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("Введите сообщение M (для B): ")
    m, _ := reader.ReadString('\n')
    m = strings.TrimSpace(m)
    if m == "" {
        fmt.Println("Ошибка: M не может быть пустым.")
        os.Exit(1)
    }

    var useTimestamp bool
    for {
        fmt.Print("Выберите тип идентификатора (1 для Timestamp, 2 для Random number): ")
        choice, _ := reader.ReadString('\n')
        choice = strings.TrimSpace(choice)
        if choice == "1" {
            useTimestamp = true
            fmt.Println("Используется Timestamp в качестве идентификатора.")
            break
        } else if choice == "2" {
            useTimestamp = false
            fmt.Println("Используется случайное число в качестве идентификатора.")
            break
        } else {
            fmt.Println("Неверный выбор. Введите 1 или 2.")
        }
    }

    
    var pubA, privA, pubB, privB map[string]interface{}
    if err := cipher.LoadFromFile(PUB_KEY_A, &pubA); err != nil {
        println("Ошибка загрузки публичного ключа A:", err.Error())
        os.Exit(1)
    }
    if err := cipher.LoadFromFile(PRIV_KEY_A, &privA); err != nil {
        println("Ошибка загрузки приватного ключа A:", err.Error())
        os.Exit(1)
    }
    if err := cipher.LoadFromFile(PUB_KEY_B, &pubB); err != nil {
        println("Ошибка загрузки публичного ключа B:", err.Error())
        os.Exit(1)
    }
    if err := cipher.LoadFromFile(PRIV_KEY_B, &privB); err != nil {
        println("Ошибка загрузки приватного ключа B:", err.Error())
        os.Exit(1)
    }

    pubAE := new(big.Int)
    pubAN := new(big.Int)
    privAD := new(big.Int)
    privAN := new(big.Int)
    pubBE := new(big.Int)
    pubBN := new(big.Int)
    privBD := new(big.Int)
    privBN := new(big.Int)
    pubAE.SetString(pubA["SubjectPublicKeyInfo"].(map[string]interface{})["publicExponent"].(string), 10)
    pubAN.SetString(pubA["SubjectPublicKeyInfo"].(map[string]interface{})["N"].(string), 10)
    privAD.SetString(privA["privateExponent"].(string), 10)
    privAN.SetString(pubAN.String(), 10)
    pubBE.SetString(pubB["SubjectPublicKeyInfo"].(map[string]interface{})["publicExponent"].(string), 10)
    pubBN.SetString(pubB["SubjectPublicKeyInfo"].(map[string]interface{})["N"].(string), 10)
    privBD.SetString(privB["privateExponent"].(string), 10)
    privBN.SetString(pubBN.String(), 10)

    go func() {
        if !RunUserB(privBD, privBN) {
            println("Сторона B завершилась с ошибкой.")
            os.Exit(1)
        }
    }()

    time.Sleep(1 * time.Second)

    if !RunUserA(m, pubBE, pubBN, useTimestamp) {
        println("Сторона A завершилась с ошибкой.")
        os.Exit(1)
    }
}