package main

import (
	//"crypto/rand"
	//"fmt"
	"math"
	"math/big"
	"math/rand"
	"time"
	"sample-app/mylib"
)

func F(x, n *big.Int) *big.Int {
	result := new(big.Int)
	result.Mul(x, x)
	result.Add(result, big.NewInt(1)) //
	result.Mod(result, n)
	return result
}
func GeneratePrimes(limit int) []*big.Int {
	primes := []*big.Int{}
	i := big.NewInt(2)
	for len(primes) < limit {
		isPrime := true
		for _, p := range primes {
			mod := new(big.Int)
			mod.Mod(i, p)
			if mod.Cmp(big.NewInt(0)) == 0 {
				isPrime = false
				break
			}
		}
		if isPrime {
			primes = append(primes, new(big.Int).Set(i))
		}
		i.Add(i, big.NewInt(1))
	}
	return primes
}

func PollardRho(n *big.Int) *big.Int {

	rand.Seed(time.Now().UnixNano())
	c := big.NewInt(int64(rand.Intn(int(n.Int64()))))
	a := c
	b := c
	for i := 1; i < 1001; i++ {
		a = F(a, n)
		b = F(F(b, n), n)
		sub := new(big.Int).Sub(a, b)
		d, _, _ := mylib.ExtendedGCD(sub.Abs(sub), n)
		if d.Cmp(big.NewInt(1)) != 0 && d.Cmp(n) != 0 && d.Cmp(n) == -1 {
			return d
		}
	}
	return nil

}

func PollardRhoMinusOne(n *big.Int) *big.Int {
	rand.Seed(time.Now().UnixNano())
	a := big.NewInt(int64(rand.Intn(100) + 2))
	d, _, _ := mylib.ExtendedGCD(a, n)

	if d.Cmp(big.NewInt(1)) != 0 {
		return d
	}

	primes := GeneratePrimes(1000) 
	for _, p := range primes {
		pBig := p
		l := int(math.Floor(math.Log(float64(n.Int64())) / math.Log(float64(p.Int64()))))
		exp := mylib.FastPowMod(pBig, big.NewInt(int64(l)), n) 
		a = mylib.FastPowMod(a, exp, n)                        // a = a^(p^l) mod n
		d, _, _ = mylib.ExtendedGCD(new(big.Int).Sub(a, big.NewInt(1)), n)
		if d.Cmp(big.NewInt(1)) != 0 && d.Cmp(n) != 0 && d.Cmp(n) == -1 {
			return d
		}

	}

	return nil
}
