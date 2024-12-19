package main

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
)

func extendedGCD(a, b *big.Int) (*big.Int, *big.Int, *big.Int) {

	old_s := big.NewInt(1)
	s := big.NewInt(0)
	old_t := big.NewInt(0)
	t := big.NewInt(1)

	for b.Cmp(big.NewInt(0)) != 0 {

		quotient := new(big.Int).Div(a, b)
		//ri+1 = ri-1 - qi*ri
		//     V
		a, b = b, new(big.Int).Mod(a, b)
		//si+1 = si-1 - qi * si
		//     V
		old_s, s = s, new(big.Int).Sub(old_s, new(big.Int).Mul(quotient, s))
		//ti+1 = ti-1 - qi * ti
		//     V
		old_t, t = t, new(big.Int).Sub(old_t, new(big.Int).Mul(quotient, t))
	}

	return a, old_s, old_t
}
func fastPow(base *big.Int, degree *big.Int) *big.Int {
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

func fastPowMod(base *big.Int, degree *big.Int, mod *big.Int) *big.Int {
	result := big.NewInt(1)
	zero := big.NewInt(0)

	for degree.Cmp(zero) > 0 {
		if degree.Bit(0) == 1 {
			result.Mul(result, base)
			result.Mod(result, mod)
		}
		base.Mul(base, base)
		base.Mod(base, mod)
		degree.Rsh(degree, 1)
	}
	return result
}

func fermatPrimalityTest(n *big.Int) bool {
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

	r := fastPowMod(a, nMinusOne, n)

	return r.Cmp(one) == 0

}

func jacobiSymbol(a *big.Int, n *big.Int) int {
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
	r := fastPowMod(a, degree, n)

	if r.Cmp(big.NewInt(1)) != 0 && r.Cmp(nMinusOne) != 0 {
		return false
	}

	s := jacobiSymbol(a, n)

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
		y := fastPowMod(a, r, n)

		if y.Cmp(big.NewInt(1)) == 0 || y.Cmp(nMinusOne) == 0 {
			continue
		}

		for j := 1; j < s; j++ {
			y = fastPowMod(y, big.NewInt(2), n)
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

func generatePrimeBitK(k int, t int) *big.Int {

	for {

		p := new(big.Int)

		p.SetBit(p, k-1, 1)
		p.SetBit(p, 0, 1)

		for i := 1; i < k-1; i++ {
			bit, _ := rand.Int(rand.Reader, big.NewInt(2))
			p.SetBit(p, i, uint(bit.Int64()))

		}
		fmt.Println(p)

		if MillerRabinPrimalityTest(p, t) {
			return p

		}

	}

}

func sloveFirstDegreeComparsion(a, b, m *big.Int) []*big.Int {

	gcd, x, _ := extendedGCD(a, m)
	if new(big.Int).Mod(b, gcd).Cmp(big.NewInt(0)) != 0 {
		fmt.Println("Нет решениий")
		return nil
	}

	b = new(big.Int).Div(b, gcd)
	m = new(big.Int).Div(m, gcd)

	x0 := new(big.Int).Mul(b, x)
	x0.Mod(x0, m)

	solutions := []*big.Int{}
	for i := big.NewInt(0); i.Cmp(gcd) < 0; i.Add(i, big.NewInt(1)) {
		solution := new(big.Int).Add(x0, new(big.Int).Mul(i, m))
		solutions = append(solutions, solution)
	}

	return solutions
}

func SolveQuadraticComparsion(a, p *big.Int) []*big.Int {

	t := 100
	if !MillerRabinPrimalityTest(p, t) {
		fmt.Println("p не является простым")
		return nil
	}

	if jacobiSymbol(a, p) != 1 {
		fmt.Println("Нет решений")
		return nil
	}

	N := new(big.Int)
	for N.SetInt64(2); jacobiSymbol(N, p) != -1; N.Add(N, big.NewInt(1)) {
	}

	_, x, _ := extendedGCD(a, p)
	k := 0
	h := new(big.Int).Sub(p, big.NewInt(1))
	for h.Bit(0) == 0 {
		h.Rsh(h, 1)
		k++
	}

	// a1 = a^((h + 1) / 2) mod p
	hPlusOne := new(big.Int).Add(h, big.NewInt(1))
	hPlusOne.Div(hPlusOne, big.NewInt(2))
	a1 := fastPowMod(a, hPlusOne, p)

	// a2 = a^(-1) mod p
	a2 := new(big.Int).Mul(a, x)
	a2.Mod(a2, p)

	// N1 = N^h mod p
	N1 := fastPowMod(N, h, p)
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
		exp := fastPow(big.NewInt(2), big.NewInt(int64(k-2-i)))
		d := fastPowMod(c, exp, p)

		ji := 0
		if d.Cmp(big.NewInt(1)) == 0 {
			ji = 0
		} else {
			ji = 1
		}

		// N2 = N2 * N1^(2^i * ji) mod p
		exp1 := fastPow(big.NewInt(2), big.NewInt(int64(i)))
		N2.Mul(N2, fastPow(N1, exp1.Mul(exp1, big.NewInt(int64(ji)))))
		N2.Mod(N2, p)
	}

	x1 := new(big.Int).Mul(a1, N2)
	x1.Mod(x1, p)

	x2 := new(big.Int).Sub(p, x1)
	return []*big.Int{x1, x2}

}

func chianeseRemainder(b, m []*big.Int) *big.Int {
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
		_, xi, _ := extendedGCD(m[i], M)
		Ni := new(big.Int).Mul(Mi, xi)
		Ni.Mod(Ni, m[i])
		prod := new(big.Int).Mul(new(big.Int).Mul(b[i], Mi), Ni)
		x.Add(x, prod)
	}
	x.Mod(x, M)
	return x

}
func NewPolynomial(coeffs []int64, mod int64) ([]*big.Int, *big.Int) {
	bigCoeffs := make([]*big.Int, len(coeffs))
	modul := big.NewInt(mod)
	for i, c := range coeffs {
		bigCoeffs[i] = new(big.Int).Mod(big.NewInt(c), modul)
	}
	return bigCoeffs, modul
}

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
func main() {
	for {

		var option int
		fmt.Println("Выберите опцию:")
		fmt.Println("1 - Алгоритм Евклида")
		fmt.Println("2 - Быстрое возведение в степень")
		fmt.Println("3 - Быстрое возведение в степень по модулю")
		fmt.Println("4 - Проверка числа на простоту по тесту Ферма")
		fmt.Println("5 - Вычисление символа Якоби")
		fmt.Println("6 - Проверка числа на простоту по тесту Соловэя-Штрассена")
		fmt.Println("7 - Проверка числа на простоту по тесту Миллера-Рабина")
		fmt.Println("8 - Генерация простого числа с разрядностью k")
		fmt.Println("9 - Решение сравнения первой степени")
		fmt.Println("10 - Решение сравнения второй степени")
		fmt.Println("11 - Решение системы сравнений")
		fmt.Println("12 - построение конечного поля")
		fmt.Println("ctrl + c - для выхода")
		fmt.Scan(&option)

		switch option {
		case 1:
			a := big.NewInt(0)
			b := big.NewInt(0)
			fmt.Println("Введите число:")
			fmt.Scan(a)
			fmt.Println("Введите число:")
			fmt.Scan(b)
			gcd, x, y := extendedGCD(a, b)
			fmt.Printf("НОД(%s, %s) = %s\n", a.String(), b.String(), gcd.String())
			fmt.Printf("Коэффициенты Безу: x = %s, y = %s\n", x.String(), y.String())
			fmt.Printf("Линейное представление НОД: %s * %s + %s * %s = %s\n", a.String(), x.String(), b.String(), y.String(), gcd.String())
		case 2:
			a := big.NewInt(0)
			b := big.NewInt(0)
			fmt.Println("Введите основание:")
			fmt.Scan(a)
			fmt.Println("Введите степень:")
			fmt.Scan(b)
			result := fastPow(a, b)
			fmt.Printf("Результат быстрого возведения числа в степень: %s\n", result.String())
		case 3:
			a := big.NewInt(0)
			b := big.NewInt(0)
			n := big.NewInt(0)
			fmt.Println("Введите основание:")
			fmt.Scan(a)
			fmt.Println("Введите степень:")
			fmt.Scan(b)
			fmt.Println("Введите модуль:")
			fmt.Scan(n)
			result := fastPowMod(a, b, n)
			fmt.Printf("Результат быстрого возведения числа в степень по модулю %s : %s\n", n.String(), result.String())
		case 4:
			n := big.NewInt(0)
			fmt.Println("Введите целое нечётное число ≥ 5:")
			fmt.Scan(n)
			if fermatPrimalityTest(n) {
				fmt.Printf("%s - вероятно число простое\n", n.String())
			} else {
				fmt.Printf("%s - число составное\n", n.String())
			}

		case 5:
			n := big.NewInt(0)
			a := big.NewInt(0)
			fmt.Println("Введите целое нечётное число n ≥ 3:")
			fmt.Scan(n)
			fmt.Println("Введите целое число 0 ≤ a < n:")
			fmt.Scan(a)
			fmt.Printf("Символ Якоби: %d\n", jacobiSymbol(a, n))

		case 6:
			n := big.NewInt(0)
			fmt.Println("Введите целое нечётное число ≥ 5:")
			fmt.Scan(n)
			if ShtrassPrimaltivityTest(n) {
				fmt.Printf("%s - вероятно число простое\n", n.String())
			} else {
				fmt.Printf("%s - число составное\n", n.String())
			}
		case 7:
			n := big.NewInt(0)
			var t int
			fmt.Println("Введите целое нечётное число ≥ 5:")
			fmt.Scan(n)
			fmt.Println("Введите число итераций:")
			fmt.Scan(&t)
			if MillerRabinPrimalityTest(n, t) {
				fmt.Printf("%s - вероятно число простое\n", n.String())
			} else {
				fmt.Printf("%s - число составное\n", n.String())
			}
		case 8:
			var k, t int
			fmt.Println("Введите разрядность k:")
			fmt.Scan(&k)
			fmt.Println("Введите количество проверок t:")
			fmt.Scan(&t)
			var probe float64 = 1 - 1/math.Pow(4, float64(t))
			p := generatePrimeBitK(k, t)
			fmt.Printf("Простое число разрядности %d: %s с вероятностью %f\n", k, p.String(), probe)
		case 9:
			a := big.NewInt(0)
			b := big.NewInt(0)
			m := big.NewInt(0)
			fmt.Println("Введите коэффициент a:")
			fmt.Scan(a)
			fmt.Println("Введите коэффициент b:")
			fmt.Scan(b)
			fmt.Println("Введите модуль m:")
			fmt.Scan(m)
			if solutions := sloveFirstDegreeComparsion(a, b, m); solutions != nil {
				for _, solution := range solutions {
					fmt.Printf("Решение: %s\n", solution.String())
				}
			}
		case 10:
			a := big.NewInt(0)
			p := big.NewInt(0)
			fmt.Println("Введите коэффициент a:")
			fmt.Scan(a)
			fmt.Println("Введите модуль p - простое:")
			fmt.Scan(p)
			if solutions := SolveQuadraticComparsion(a, p); solutions != nil {
				for _, solution := range solutions {
					fmt.Printf("Решение: %s\n", solution.String())
				}
			}
		case 11:
			var n int
			b := []*big.Int{}
			m := []*big.Int{}
			fmt.Println("Введите количество сравнений:")
			fmt.Scan(&n)
			for i := 0; i < n; i++ {
				bi := big.NewInt(0)
				mi := big.NewInt(0)
				fmt.Println("Введите число b:")
				fmt.Scan(bi)
				b = append(b, bi)
				fmt.Println("Введите число m:")
				fmt.Scan(mi)
				m = append(m, mi)
			}
			soultion := chianeseRemainder(b, m)
			if soultion != nil {
				fmt.Printf("Решение: %s\n", soultion.String())
			} else {
				fmt.Println("Нет решения")
			}
		case 12:
			mod := big.NewInt(2)

			// x^2 + x + 1
			modPoly := []*big.Int{big.NewInt(1), big.NewInt(1), big.NewInt(1)}
			fmt.Println("Неприводимый: ")
			PrintPolynomial(modPoly)
			// x + 1
			p1 := []*big.Int{big.NewInt(1), big.NewInt(1)}
			p2 := []*big.Int{big.NewInt(1), big.NewInt(1)}
			fmt.Println("p1: ")
			PrintPolynomial(p1)
			fmt.Println("p2: ")
			PrintPolynomial(p2)
			sum := Add(p1, p2, mod)
			fmt.Println("Сумма: ")
			PrintPolynomial(sum)

			prod := Mul(p1, p2, mod) // [1, 0, 1]
			fmt.Println("Результат приведения: ")
			// Приведение по модулю неприводимого многочлена
			result := Mod(prod, modPoly, mod)
			// [0, 1]
			PrintPolynomial(result)
		default:
			fmt.Printf("Неверный вариант")
		}

	}

}
