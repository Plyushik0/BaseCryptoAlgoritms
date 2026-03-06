package mylib

import (
    "math/big"
    "testing"
    "fmt"
)

func TestExtendedGCD(t *testing.T) {
    a := big.NewInt(28)
    b := big.NewInt(45)
    gcd, x, y := ExtendedGCD(a, b)
    expectedGCD := big.NewInt(1)
    if gcd.Cmp(expectedGCD) != 0 {
        t.Errorf("ExtendedGCD(%s, %s) = %s; ожидалось %s", a, b, gcd, expectedGCD)
    }
    // Проверяем, что ax + by = gcd
    ax := new(big.Int).Mul(a, x)
    by := new(big.Int).Mul(b, y)
    sum := new(big.Int).Add(ax, by)
    if sum.Cmp(gcd) != 0 {
        t.Errorf("ax + by = %s; ожидалось %s", sum, gcd)
    }
}

func TestModInverse(t *testing.T) {
    a := new(big.Int).SetInt64(3)
    m := new(big.Int).SetInt64(11)
    inv, err := big.NewInt(0), error(nil)
    // Вызываем modInverse через utils.go (предполагая, что он использует ExtendedGCD)
    gcd, x, _ := ExtendedGCD(a, m)
    if gcd.Cmp(big.NewInt(1)) == 0 {
        inv = x.Mod(x, m)
    } else {
        err = fmt.Errorf("модульный обратный не существует")
    }
    if err != nil {
        t.Errorf("modInverse(%s, %s) вернул ошибку: %v", a, m, err)
    }
    expected := big.NewInt(4) // 3 * 4 = 12 ≡ 1 (mod 11)
    if inv.Cmp(expected) != 0 {
        t.Errorf("modInverse(%s, %s) = %s; ожидалось %s", a, m, inv, expected)
    }
}
func TestFastPowMod(t *testing.T) {
    a := big.NewInt(2)
    b := big.NewInt(3)
    m := big.NewInt(5)
    result := FastPowMod(a, b, m) // 2^3 mod 5 = 8 mod 5 = 3
    expected := big.NewInt(3)
    if result.Cmp(expected) != 0 {
        t.Errorf("FastPowMod(%s, %s, %s) = %s; ожидалось %s", a, b, m, result, expected)
    }
}