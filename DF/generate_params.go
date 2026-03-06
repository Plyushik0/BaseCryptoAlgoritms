package main

import (
	"encoding/json"
	"math/big"
	"sample-app/mylib"
	"fmt"
	"os"
)


func generateAndSaveParams() {

	p := mylib.GeneratePrimeBitK(PRIME_BITS, 100)

	g := findPrimitiveRoot(p)
	
	params := map[string]string{
		"p": p.String(),
		"g": g.String(),
	}
	
	data, err := json.Marshal(params)
	if err != nil {
		fmt.Printf("Ошибка сериализации параметров: %v\n", err)
		return
	}
	if err := os.WriteFile(PARAMS_FILE, data, 0644); err != nil {
		fmt.Printf("Ошибка записи в файл %s: %v\n", PARAMS_FILE, err)
		return
	}
	fmt.Printf("Параметры сохранены в %s\n", PARAMS_FILE)
}


func findPrimitiveRoot(p *big.Int) *big.Int {
	α := big.NewInt(2)
	one := big.NewInt(1)
	pm1 := new(big.Int).Sub(p, one)
	exp := new(big.Int).Div(pm1, big.NewInt(2))

	for {
		pow := mylib.FastPowMod(α, exp, p)
		if pow.Cmp(one) != 0 {
			return α
		}
		α.Add(α, one)
	}
}




func clearFile() {
	if err := os.WriteFile(PARAMS_FILE, []byte{}, 0644); err != nil {
		fmt.Printf("Ошибка очистки файла %s: %v\n", PARAMS_FILE, err)
	}
}

func main() {
	clearFile()
	generateAndSaveParams()
}