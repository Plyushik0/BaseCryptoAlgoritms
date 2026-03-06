package main

import (
	"encoding/hex"

	"fmt"
	"sample-app/hashf" 
)


func HashMessage(message []byte, hashFuncName string) ([]byte, error) {
	switch hashFuncName {
	case "sha256":
		result := hashf.Sha256(message)
		return result[:], nil
	case "sha512":
		result := hashf.Sha512(message)
		return result[:], nil
	case "streebog512":
		result := hashf.StreebogHash(message, 64)
		return result[:], nil
	default:
		return nil, fmt.Errorf("неподдерживаемая хеш-функция: %s", hashFuncName)
	}
}


func IterateHash(value []byte, k int, hashFuncName string) ([]byte, error) {
	currentValue := value
	for i := 0; i < k; i++ {
		var err error
		currentValue, err = HashMessage(currentValue, hashFuncName)
		if err != nil {
			return nil, fmt.Errorf("ошибка при итерации %d: %v", i, err)
		}
	}
	return currentValue, nil
}


func BytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}


func HexToBytes(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}