package main

import (
	// "encoding/json"
	"math/big"
	"sample-app/mylib"
	// "sample-app/mylib"
	//"crypto/rand"
)


func GenerateParameters(bitSize int) (*big.Int, *big.Int, *big.Int) {
	var p, q *big.Int

	
	for {
		p = mylib.GeneratePrimeBitK(bitSize, 100)
		if new(big.Int).Mod(p, big.NewInt(4)).Cmp(big.NewInt(3)) == 0 {
			break
		}
	}

	
	for {
		q = mylib.GeneratePrimeBitK(bitSize, 100)
		if new(big.Int).Mod(q, big.NewInt(4)).Cmp(big.NewInt(3)) == 0 && p.Cmp(q) != 0 {
			break
		}
	}
	
	n := new(big.Int).Mul(p,q)
	return p, q, n
}


func encryptRabin(message *big.Int, n *big.Int) *big.Int {
	
	c := mylib.FastPowMod(message, big.NewInt(2), n)
	return c
}




func decryptRabin(ciphertext *big.Int, p *big.Int, q *big.Int) (*big.Int, *big.Int, *big.Int, *big.Int) {
	// mp = c^((p+1)/4) mod p
	expP := new(big.Int).Add(p, big.NewInt(1))
	expP.Div(expP, big.NewInt(4))
	mp := mylib.FastPowMod(ciphertext, expP, p)

	// mq = c^((q+1)/4) mod q
	expQ := new(big.Int).Add(q, big.NewInt(1))
	expQ.Div(expQ, big.NewInt(4))
	mq := mylib.FastPowMod(ciphertext, expQ, q)
	
	n := new(big.Int).Mul(p, q)
	yp := new(big.Int)
	yq := new(big.Int)
	
	
	
	_, yp, yq = mylib.ExtendedGCD(p, q)

	
	r1 := new(big.Int).Mul(new(big.Int).Mul(mp, q), yq)
	r2 := new(big.Int).Mul(new(big.Int).Mul(mq, p), yp)
	r1.Add(r1, r2).Mod(r1, n)

	r3 := new(big.Int).Sub(n, r1)
	r4 := new(big.Int).Mul(new(big.Int).Mul(mp, q), yq)
	r5 := new(big.Int).Mul(new(big.Int).Mul(mq, p), yp)
	r4.Sub(r4, r5).Mod(r4, n)

	r6 := new(big.Int).Sub(n, r4)

	return r1, r3, r4, r6
}








