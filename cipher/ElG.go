package cipher

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sample-app/mylib"
)

// Генерация ключей
func GenerateELGKeys(bitSize int, pubPath, privPath string) (*big.Int, *big.Int, *big.Int, *big.Int) {
    p := mylib.GeneratePrimeBitK(bitSize, 100)
	α := big.NewInt(2)
    one := big.NewInt(1)
    pm1 := new(big.Int).Sub(p, one)
    exp := new(big.Int).Div(pm1, big.NewInt(2))
	
    for {
        pow := mylib.FastPowMod(α, exp, p)
        if pow.Cmp(one) != 0 {
            break
        }
        α.Add(α, one)
    }
    a, _ := rand.Int(rand.Reader, p)
    β := mylib.FastPowMod(α, a, p)


    pubKey := map[string]interface{}{
        "SubjectPublicKeyInfo": map[string]interface{}{
            "alpha": α.String(),
            "beta":  β.String(),
            "p":     p.String(),
        },
    }
    SaveToFile(pubPath, pubKey)

    privKey := map[string]interface{}{
        "privateExponent": a.String(),
    }
    SaveToFile(privPath, privKey)

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