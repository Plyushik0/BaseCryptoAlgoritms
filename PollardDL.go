package main

import (
	//"fmt"
	"math/big"
	"sample-app/mylib"
	// "math/rand"
	// "time"
	// "os"
)


func NewF(c [2]*big.Int, a, b, p *big.Int) [2]*big.Int {
	cN := new(big.Int).Mul(mylib.FastPowMod(a, c[0], p), mylib.FastPowMod(b, c[1], p))
	cN.Mod(cN, p)

	pHalf := new(big.Int).Div(p, big.NewInt(2))

	if pHalf.Cmp(cN) > 0 {
		c[0].Add(c[0], big.NewInt(1))
		c[0].Mod(c[0], new(big.Int).Sub(p, big.NewInt(1)))
	} else {
		c[1].Add(c[1], big.NewInt(1))
		c[1].Mod(c[1], new(big.Int).Sub(p, big.NewInt(1)))
	}

	return c
}

func PollardDL(u, v, a, b, p *big.Int) *big.Int {
	c := [2]*big.Int{u, v}
	d := [2]*big.Int{new(big.Int).Set(u), new(big.Int).Set(v)}

	for i := 0; i < 1000; i++ {
		c = NewF(c, a, b, p)
		d = NewF(NewF(d, a, b, p), a, b, p)

		cN := new(big.Int).Mul(mylib.FastPow(a, c[0]), mylib.FastPow(b, c[1]))
		cN.Mod(cN, p)

		dN := new(big.Int).Mul(mylib.FastPow(a, d[0]), mylib.FastPow(b, d[1]))
		dN.Mod(dN, p)

		if cN.Cmp(dN) == 0 {
			aC := new(big.Int).Sub(c[1], d[1])
			aC.Mod(aC, new(big.Int).Sub(p, big.NewInt(1)))

			bC := new(big.Int).Sub(d[0], c[0])
			bC.Mod(bC, new(big.Int).Sub(p, big.NewInt(1)))

			solutions := mylib.SloveFirstDegreeComparsion(aC, bC, new(big.Int).Sub(p, big.NewInt(1)))
			return solutions
		}
	}

	return nil
}