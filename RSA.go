package main

import (

	"encoding/json"
	"sample-app/mylib"
	"math/big"
	"os"
	"crypto/rand"
)




func saveToFile(filename string, data interface{}) error {
	file, _ := json.MarshalIndent(data, "", "  ")
	return os.WriteFile(filename, file, 0644)
}

func loadFromFile(filename string, data interface{}) error {
	file, _ := os.ReadFile(filename)
	return json.Unmarshal(file, data)
}


func generateRSAKeys(bitSize int) (e, d, n, fn *big.Int) {
	p := mylib.GeneratePrimeBitK(bitSize, 20)
	q := mylib.GeneratePrimeBitK(bitSize, 20)

	n = new(big.Int).Mul(p, q)
	fn = new(big.Int).Mul(new(big.Int).Sub(p, big.NewInt(1)), new(big.Int).Sub(q, big.NewInt(1)))


	for {
		e, _ = rand.Int(rand.Reader, fn)
		nod, _, _ := mylib.ExtendedGCD(e, fn)
		if e.Cmp(big.NewInt(1)) > 0 && nod.Cmp(big.NewInt(1)) == 0 {
			break
		}
	}


	d = mylib.SloveFirstDegreeComparsion(e, big.NewInt(1), fn)
	return e, d, n, fn
}

// PKCS 7
func encryptRSA(message string, e, n *big.Int) []*big.Int {
	
	M := []byte(message)
	// Вычисляем padding
	blockSize := 254 
	padding := blockSize - (len(M) % blockSize)

	// Добавляем padding
	paddedMessage := append(M, byte(padding))
	for i := 1; i < padding; i++ {
		paddedMessage = append(paddedMessage, byte(padding))
	}

	// Разбиваем на блоки и шифруем
	var C []*big.Int
	for i := 0; i < len(paddedMessage); i += blockSize {
		block := paddedMessage[i : i+blockSize]
		m := new(big.Int).SetBytes(block)
		c := mylib.FastPowMod(m, e, n)
		C = append(C, c)
	}

	return C
}


func decryptRSA(C []*big.Int, d, n *big.Int) string {
	var decryptedBytes []byte

	// Дешифруем каждый блок
	for _, c := range C {
		m := mylib.FastPowMod(c, d, n)
		decryptedBytes = append(decryptedBytes, m.Bytes()...)
	}

	// Удаляем padding
	padding := int(decryptedBytes[len(decryptedBytes)-1])
	if padding > len(decryptedBytes) {
		return "" // Некорректный padding
	}
	return string(decryptedBytes[:len(decryptedBytes)-padding])
}