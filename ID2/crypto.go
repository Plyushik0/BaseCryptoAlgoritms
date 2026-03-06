package main

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "sample-app/hashf"
)

func EncryptAES(key, data []byte) (iv, ciphertext []byte) {
    if len(key) != KEY_SIZE {
        println("Неверный размер ключа")
        return nil, nil
    }

    hash := hashf.Sha256(data)
    dataToEncrypt := append(data, hash[:]...)

    block, err := aes.NewCipher(key)
    if err != nil {
        println("Ошибка создания AES-шифра:", err.Error())
        return nil, nil
    }

    iv = make([]byte, aes.BlockSize)
    if _, err := rand.Read(iv); err != nil {
        println("Ошибка генерации IV:", err.Error())
        return nil, nil
    }

    paddedData := pad(dataToEncrypt, aes.BlockSize)
    ciphertext = make([]byte, len(paddedData))
    mode := cipher.NewCBCEncrypter(block, iv)
    mode.CryptBlocks(ciphertext, paddedData)

    return iv, ciphertext
}

func DecryptAES(key, iv, ciphertext []byte) []byte {
    if len(key) != KEY_SIZE {
        println("Неверный размер ключа")
        return nil
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        println("Ошибка создания AES-шифра:", err.Error())
        return nil
    }

    if len(ciphertext)%aes.BlockSize != 0 {
        println("Неверный размер зашифрованных данных")
        return nil
    }

    plaintext := make([]byte, len(ciphertext))
    mode := cipher.NewCBCDecrypter(block, iv)
    mode.CryptBlocks(plaintext, ciphertext)

    unpadded := unpad(plaintext)
    if len(unpadded) < 32 {
        println("Неверный формат расшифрованных данных")
        return nil
    }

    data := unpadded[:len(unpadded)-32]
    receivedHash := unpadded[len(unpadded)-32:]
    expectedHash := hashf.Sha256(data)

    if string(receivedHash) != string(expectedHash[:]) {
        println("Хэш данных не совпадает. Данные повреждены.")
        return nil
    }

    return data
}

func pad(data []byte, blockSize int) []byte {
    padding := blockSize - len(data)%blockSize
    padText := make([]byte, len(data)+padding)
    copy(padText, data)
    for i := len(data); i < len(padText); i++ {
        padText[i] = byte(padding)
    }
    return padText
}

func unpad(data []byte) []byte {
    if len(data) == 0 {
        return nil
    }
    padding := int(data[len(data)-1])
    if padding > len(data) || padding == 0 {
        return nil
    }
    return data[:len(data)-padding]
}