package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sample-app/mylib"
)

// Генерация ключей
func generateKeys(bitSize int) (*big.Int, *big.Int, *big.Int, *big.Int) {

	p := mylib.GeneratePrimeBitK(bitSize, 100)
	α := big.NewInt(2)
	// pMinusTwo := new(big.Int).Sub(p, big.NewInt(2))
	a, _ := rand.Int(rand.Reader, p)

	β := mylib.FastPowMod(α, a, p)

	return p, α, β, a
}


func encryptELG(message, p, α, β *big.Int) (*big.Int, *big.Int) { 
	r, _ := rand.Int(rand.Reader, p)
	c1 := mylib.FastPowMod(α, r, p)
	c2 := mylib.FastPowMod(β, r, p)
	c2.Mul(c2, message).Mod(c2, p)

	return c1, c2
}


func decryptELG(c1, c2, a, p *big.Int) *big.Int {
	fmt.Println(a)
	// s = c1^a mod p
	s := mylib.FastPowMod(c1, a, p)

	// s^-1 mod p (обратный элемент для s)
	_, x, _ := mylib.ExtendedGCD(s, p)
	x.Mod(x, p)
	// c2 * s^-1 mod p
	message := new(big.Int).Mul(c2, x)
	message.Mod(message, p)

	return message
}