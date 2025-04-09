package mylib
import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func ExtendedGCD(a, b *big.Int) (*big.Int, *big.Int, *big.Int) {
	if b.Cmp(big.NewInt(0)) == 0 {
		return a, big.NewInt(1), big.NewInt(0)
	}
	gcd, x1, y1 := ExtendedGCD(b, new(big.Int).Mod(a, b))
	x := y1
	y := new(big.Int).Sub(x1, new(big.Int).Mul(y1, new(big.Int).Div(a, b)))
	return gcd, x, y
}
func FastPow(base *big.Int, degree *big.Int) *big.Int {
	result := big.NewInt(1)
	zero := big.NewInt(0)

	//two := big.NewInt(2)

	for degree.Cmp(zero) > 0 {
		if degree.Bit(0) == 1 {
			result.Mul(result, base)
		}
		base.Mul(base, base)

		//degree.Div(degree, two)
		degree.Rsh(degree, 1)
	}
	return result
}

func FastPowMod(a *big.Int, b *big.Int, m *big.Int) *big.Int {
	result := big.NewInt(1) 
	a = new(big.Int).Set(a) 
    b = new(big.Int).Set(b) 
	
	for b.Cmp(big.NewInt(0)) > 0 {
		
		if new(big.Int).And(b, big.NewInt(1)).Cmp(big.NewInt(1)) == 0 {
			result.Mul(result, a).Mod(result, m)
		}
		
		a.Mul(a, a).Mod(a, m)
		
		b.Rsh(b, 1)
	}

	return result
}

func FermatPrimalityTest(n *big.Int) bool {
	if n.Cmp(big.NewInt(0)) <= 0 || n.Bit(0) == 0 {
		panic("n должно быть положительным нечётным числом")
	}

	one := big.NewInt(1)
	two := big.NewInt(2)
	//nMinusTwo := new(big.Int).Sub(n, two)
	nMinusOne := new(big.Int).Sub(n, one)
	nMinusFour := new(big.Int).Sub(n, big.NewInt(4))
	a, _ := rand.Int(rand.Reader, nMinusFour)
	a.Add(a, two)

	r := FastPowMod(a, nMinusOne, n)

	return r.Cmp(one) == 0

}

func YacobiSymbol(a *big.Int, n *big.Int) int {
	if n.Cmp(big.NewInt(0)) <= 0 || n.Bit(0) == 0 {
		panic("n должно быть положительным нечётным числом")
	}

	g := 1
	zero := big.NewInt(0)
	one := big.NewInt(1)
	eight := big.NewInt(8)

	a = new(big.Int).Mod(a, n)

	for a.Cmp(zero) != 0 {

		if a.Cmp(zero) == 0 {
			return 0
		}

		if a.Cmp(one) == 0 {
			return g
		}

		k := 0
		for a.Bit(0) == 0 {
			a.Rsh(a, 1)
			k++
		}

		if k%2 == 1 {
			nMod8 := new(big.Int).Mod(n, eight).Int64()
			if nMod8 == 3 || nMod8 == 5 {
				g = -g
			}
		}

		if a.Bit(0) == 1 && n.Bit(0) == 1 {
			aMod4 := new(big.Int).Mod(a, big.NewInt(4))
			nMod4 := new(big.Int).Mod(n, big.NewInt(4))
			if aMod4.Cmp(big.NewInt(3)) == 0 && nMod4.Cmp(big.NewInt(3)) == 0 {
				g = -g
			}
		}

		a, n = n, a
		a = new(big.Int).Mod(a, n)
	}

	if n.Cmp(one) == 0 {
		return g
	}

	return 0
}

func ShtrassPrimaltivityTest(n *big.Int) bool {
	if n.Cmp(big.NewInt(5)) < 0 || n.Bit(0) == 0 {
		panic("n должно быть положительным нечётным числом больше 5")
	}

	//nMinusTwo := new(big.Int).Sub(n, big.NewInt(2))
	nMinusOne := new(big.Int).Sub(n, big.NewInt(1))
	nMinusFour := new(big.Int).Sub(n, big.NewInt(4))
	a, _ := rand.Int(rand.Reader, nMinusFour)

	degree := new(big.Int).Div(nMinusOne, big.NewInt(2))
	r := FastPowMod(a, degree, n)

	if r.Cmp(big.NewInt(1)) != 0 && r.Cmp(nMinusOne) != 0 {
		return false
	}

	s := YacobiSymbol(a, n)

	if s == 0 {
		return false
	}

	sMod := new(big.Int).Mod(big.NewInt(int64(s)), n)
	return r.Cmp(sMod) != 0

}

func MillerRabinPrimalityTest(n *big.Int, t int) bool {
	if n.Cmp(big.NewInt(5)) < 0 || n.Bit(0) == 0 {
		panic("n должно быть положительным нечётным числом больше 5")
	}

	nMinusOne := new(big.Int).Sub(n, big.NewInt(1))
	//nMinusTwo := new(big.Int).Sub(n, big.NewInt(2))
	nMinusFour := new(big.Int).Sub(n, big.NewInt(4))

	r := new(big.Int).Set(nMinusOne)
	s := 0
	for r.Bit(0) == 0 {
		r.Rsh(r, 1)
		s++
	}

	for i := 0; i < t; i++ {

		a, _ := rand.Int(rand.Reader, nMinusFour)

		a.Add(a, big.NewInt(2))
		y := FastPowMod(a, r, n)

		if y.Cmp(big.NewInt(1)) == 0 || y.Cmp(nMinusOne) == 0 {
			continue
		}

		for j := 1; j < s; j++ {
			y = FastPowMod(y, big.NewInt(2), n)
			if y.Cmp(big.NewInt(1)) == 0 {
				return false
			}
			if y.Cmp(nMinusOne) == 0 {
				break
			}
		}
		if y.Cmp(nMinusOne) != 0 {
			return false
		}

	}
	return true

}

func GeneratePrimeBitK(k int, t int) *big.Int {

	for {

		p := new(big.Int)

		p.SetBit(p, k-1, 1)
		p.SetBit(p, 0, 1)

		for i := 1; i < k-1; i++ {
			bit, _ := rand.Int(rand.Reader, big.NewInt(2))
			p.SetBit(p, i, uint(bit.Int64()))

		}
		
		if MillerRabinPrimalityTest(p, t) {
			return p

		}

	}

}

func SloveFirstDegreeComparsion(a, b, m *big.Int) *big.Int {

	gcd, x, _ := ExtendedGCD(a, m)
	if new(big.Int).Mod(b, gcd).Cmp(big.NewInt(0)) != 0 {
		fmt.Println("Нет решениий")
		return nil
	}

	b = new(big.Int).Div(b, gcd)
	m = new(big.Int).Div(m, gcd)

	x0 := new(big.Int).Mul(b, x)
	x0.Mod(x0, m)

	

	return x0
}

func SolveQuadraticComparsion(a, p *big.Int) *big.Int {

	t := 100
	if !MillerRabinPrimalityTest(p, t) {
		fmt.Println("p не является простым")
		return nil
	}

	if YacobiSymbol(a, p) != 1 {
		fmt.Println("Нет решений")
		return nil
	}

	N := new(big.Int)
	for N.SetInt64(2); YacobiSymbol(N, p) != -1; N.Add(N, big.NewInt(1)) {
	}

	_, x, _ := ExtendedGCD(a, p)
	k := 0
	h := new(big.Int).Sub(p, big.NewInt(1))
	for h.Bit(0) == 0 {
		h.Rsh(h, 1)
		k++
	}

	// a1 = a^((h + 1) / 2) mod p
	hPlusOne := new(big.Int).Add(h, big.NewInt(1))
	hPlusOne.Div(hPlusOne, big.NewInt(2))
	a1 := FastPowMod(a, hPlusOne, p)

	// a2 = a^(-1) mod p
	a2 := new(big.Int).Mul(a, x)
	a2.Mod(a2, p)

	// N1 = N^h mod p
	N1 := FastPowMod(N, h, p)
	N2 := big.NewInt(1) // Начальное значение N2

	for i := 0; i < k-2; i++ {
		// b = a1 * N2 mod p
		b := new(big.Int).Mul(a1, N2)
		b.Mod(b, p)

		// c = a2 * b^2 mod p
		b.Mul(b, b)
		c := new(big.Int).Mul(a2, b)
		c.Mod(c, p)

		// d = c^(2^(k-2-i)) mod p
		exp := FastPow(big.NewInt(2), big.NewInt(int64(k-2-i)))
		d := FastPowMod(c, exp, p)

		ji := 0
		if d.Cmp(big.NewInt(1)) == 0 {
			ji = 0
		} else {
			ji = 1
		}

		// N2 = N2 * N1^(2^i * ji) mod p
		exp1 := FastPow(big.NewInt(2), big.NewInt(int64(i)))
		N2.Mul(N2, FastPow(N1, exp1.Mul(exp1, big.NewInt(int64(ji)))))
		N2.Mod(N2, p)
	}

	x1 := new(big.Int).Mul(a1, N2)
	x1.Mod(x1, p)

	return x1

}

func ChianeseRemainder(b, m []*big.Int) *big.Int {
	if len(b) != len(m) {
		panic("Размеры массивов не совпадают")
	}

	M := big.NewInt(1)
	for _, mi := range m {
		M.Mul(M, mi)
	}

	x := big.NewInt(0)
	for i := 0; i < len(b); i++ {
		Mi := new(big.Int).Div(M, m[i])
		_, xi, _ := ExtendedGCD(m[i], M)
		Ni := new(big.Int).Mul(Mi, xi)
		Ni.Mod(Ni, m[i])
		prod := new(big.Int).Mul(new(big.Int).Mul(b[i], Mi), Ni)
		x.Add(x, prod)
	}
	x.Mod(x, M)
	return x

}
// func NewPolynomial(coeffs []int64, mod int64) ([]*big.Int, *big.Int) {
// 	bigCoeffs := make([]*big.Int, len(coeffs))
// 	modul := big.NewInt(mod)
// 	for i, c := range coeffs {
// 		bigCoeffs[i] = new(big.Int).Mod(big.NewInt(c), modul)
// 	}
// 	return bigCoeffs, modul
// }

func Add(p1, p2 []*big.Int, mod *big.Int) []*big.Int {
	maxLen := len(p1)
	if len(p2) > maxLen {
		maxLen = len(p2)
	}
	result := make([]*big.Int, maxLen)
	for i := 0; i < maxLen; i++ {
		var a, b *big.Int
		if i < len(p1) {
			a = p1[i]
		} else {
			a = big.NewInt(0)
		}
		if i < len(p2) {
			b = p2[i]
		} else {
			b = big.NewInt(0)
		}
		result[i] = new(big.Int).Mod(new(big.Int).Add(a, b), mod)
	}
	return result
}

func Mul(p1, p2 []*big.Int, mod *big.Int) []*big.Int {
	result := make([]*big.Int, len(p1)+len(p2)-1)
	for i := range result {
		result[i] = big.NewInt(0)
	}
	for i, a := range p1 {
		for j, b := range p2 {
			product := new(big.Int).Mod(new(big.Int).Mul(a, b), mod)
			result[i+j].Add(result[i+j], product).Mod(result[i+j], mod)
		}
	}
	return result
}

func Mod(p, modPoly []*big.Int, mod *big.Int) []*big.Int {
	result := make([]*big.Int, len(p)) // Копируем многочлен
	copy(result, p)
	for len(result) >= len(modPoly) {
		if result[len(result)-1].Cmp(big.NewInt(0)) == 0 { // Определяем страший коэффициент
			result = result[:len(result)-1] // Убираем страший нулевой коэффициент
			continue
		}
		factor := new(big.Int).Mod(result[len(result)-1], mod) // определяем множитель для вычититания
		shift := len(result) - len(modPoly)                    // определяем сдвиг степени
		for i, c := range modPoly {                            // вычитаем по сдвигу, с - коэффициент
			index := i + shift
			term := new(big.Int).Mul(c, factor) // вычитаемое
			result[index].Sub(result[index], term).Mod(result[index], mod)
		}
		result = result[:len(result)-1]
	}
	return result
}

func PrintPolynomial(p []*big.Int) {
	for i := len(p) - 1; i >= 0; i-- {
		fmt.Printf("%dx^%d ", p[i], i)
		if i > 0 {
			fmt.Print("+ ")
		}
	}
	fmt.Println()
}


