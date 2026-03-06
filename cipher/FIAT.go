package cipher

import (


	"sample-app/mylib"
	"math/big"

	"crypto/rand"
)

func GenerateFiatKeys(m, bits int, pubFile, privFile string) {
	p := mylib.GeneratePrimeBitK(bits, 100)
	q := mylib.GeneratePrimeBitK(bits, 100)
	n := new(big.Int).Mul(p, q)

	A := make([]string, m)
	B := make([]string, m)
	for i := 0; i < m; i++ {
		var a *big.Int
		for {
			a, _ = rand.Int(rand.Reader, n)
			if a.Cmp(big.NewInt(1)) > 0 {
				gcd, _, _ := mylib.ExtendedGCD(a, n)
				if gcd.Cmp(big.NewInt(1)) == 0 {
					break
				}
			}
		}
		gcd, ainv, _ := mylib.ExtendedGCD(a, n)
		if gcd.Cmp(big.NewInt(1)) != 0 {
			panic("Обратного элемента не существует")
		}
		ainv.Mod(ainv, n)
		b := mylib.FastPowMod(ainv, big.NewInt(2), n)
		A[i] = a.String()
		B[i] = b.String()
	}
	pubKey := map[string]interface{}{
		"b": B,
		"n": n.String(),
	}
	SaveToFile(pubFile, pubKey)

	privKey := map[string]interface{}{
		"a": A,
		"p": p.String(),
		"q": q.String(),
		"n": n.String(),
	}
	SaveToFile(privFile, privKey)
}