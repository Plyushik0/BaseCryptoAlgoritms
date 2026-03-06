package main

import (
    "crypto/rand"
    "os"
)

func GenerateAndSaveKey() error {
    if _, err := os.Stat(KEY_FILE); !os.IsNotExist(err) {
        println("Предупреждение: Файл ключа уже существует. Перезаписываем.")
    }

    key := make([]byte, KEY_SIZE)
    if _, err := rand.Read(key); err != nil {
        println("Ошибка генерации ключа:", err.Error())
        return err
    }

    if err := os.WriteFile(KEY_FILE, key, 0644); err != nil {
        println("Ошибка записи ключа в файл:", err.Error())
        return err
    }

    println("Ключ успешно сгенерирован и сохранён в", KEY_FILE)
    return nil
}