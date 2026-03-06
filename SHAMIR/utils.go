package main

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"sample-app/mylib"
)


func modularInverse(a, m *big.Int) (*big.Int, error) {
	gcd, x, _ := mylib.ExtendedGCD(new(big.Int).Set(a), new(big.Int).Set(m))
	if gcd.Cmp(big.NewInt(1)) != 0 {
		return nil, fmt.Errorf("обратный элемент не существует для %v mod %v", a, m)
	}
	if x.Cmp(big.NewInt(0)) < 0 {
		x.Add(x, m)
	}
	return x.Mod(x, m), nil
}

//  вычисляет f(x) для полинома с коэффициентами coeffs по модулю prime
func evaluatePolynomial(coeffs []*big.Int, xVal, prime *big.Int) *big.Int {
	result := big.NewInt(0)
	powerOfX := big.NewInt(1)
	for _, coeff := range coeffs {
		term := new(big.Int).Mul(coeff, powerOfX).Mod(new(big.Int).Mul(coeff, powerOfX), prime)
		result.Add(result, term).Mod(result, prime)
		powerOfX.Mul(powerOfX, xVal).Mod(powerOfX, prime)
	}
	return result
}

// генерирует случайные коэффициенты полинома
func generateRandomCoefficients(degree int, prime, secret *big.Int) ([]*big.Int, error) {
	if secret.Cmp(big.NewInt(0)) < 0 || secret.Cmp(prime) >= 0 {
		return nil, fmt.Errorf("секрет %v должен быть в диапазоне [0, %v)", secret, prime)
	}
	coeffs := make([]*big.Int, degree+1)
	coeffs[0] = new(big.Int).Set(secret) // a_0 = secret
	for i := 1; i <= degree; i++ {
		coeff, err := rand.Int(rand.Reader, prime)
		if err != nil {
			return nil, fmt.Errorf("ошибка генерации случайного коэффициента: %v", err)
		}
		coeffs[i] = coeff
	}
	return coeffs, nil
}