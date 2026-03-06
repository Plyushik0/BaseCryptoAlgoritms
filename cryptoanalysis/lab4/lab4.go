package main

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strings"
)

type kv struct {
	Key   int
	Value float64
}

func main() {

	textBytes, _ := os.ReadFile("5.txt")
	text := string(textBytes)

	r1Bytes, _ := os.ReadFile("1.txt")

	fmt.Println(string(r1Bytes))

	reg := regexp.MustCompile("[^А-Я]")
	text = reg.ReplaceAllString(text, "")

	theor := map[rune]float64{
		' ': 0.175, 'О': 0.090, 'Е': 0.072, 'Ё': 0.072, 'А': 0.062,
		'И': 0.062, 'Т': 0.053, 'Н': 0.053, 'С': 0.045, 'Р': 0.040,
		'В': 0.038, 'Л': 0.035, 'К': 0.028, 'М': 0.026, 'Д': 0.025,
		'П': 0.023, 'У': 0.021, 'Я': 0.018, 'Ы': 0.016, 'З': 0.016,
		'Ь': 0.014, 'Ъ': 0.014, 'Б': 0.014, 'Г': 0.013, 'Ч': 0.012,
		'Й': 0.010, 'Х': 0.009, 'Ж': 0.007, 'Ю': 0.006, 'Ш': 0.006,
		'Ц': 0.004, 'Щ': 0.004, 'Э': 0.003, 'Ф': 0.002,
	}

	alphabetRunes := []rune("АБВГДЕЖЗИКЛМНОПРСТУФХЦЧШЩЬЫЭЮЯ")
	S := len(alphabetRunes)
	n := len([]rune(text))

	alphabetIndexMap := make(map[rune]int)
	for i, r := range alphabetRunes {
		alphabetIndexMap[r] = i
	}

	keylen := 17
	var key []int
	chi2res := make(map[int]float64)
	textRunes := []rune(text)

	for R := 2; R <= 50; R++ {
		columns := make([][]rune, R)
		for i, char := range textRunes {
			columns[i%R] = append(columns[i%R], char)
		}

		freqs := make([][]int, R)
		for i := 0; i < R; i++ {
			freqs[i] = make([]int, S)
			colStr := string(columns[i])
			for j, char := range alphabetRunes {
				freqs[i][j] = strings.Count(colStr, string(char))
			}
		}

		v_i := make([]float64, R)
		for i := 0; i < R; i++ {
			v_i[i] = float64(len(columns[i]))
		}

		v_j := make([]float64, S)
		for j := 0; j < S; j++ {
			for i := 0; i < R; i++ {
				v_j[j] += float64(freqs[i][j])
			}
		}

		chi2 := 0.0
		for i := 0; i < R; i++ {
			for j := 0; j < S; j++ {
				if v_j[j]*v_i[i] != 0 {
					chi2 += math.Pow(float64(freqs[i][j]), 2) / (v_j[j] * v_i[i])
				}
			}
		}
		chi2 = (chi2 - 1) * float64(n)
		chi2res[R] = chi2

		if R == keylen {
			fmt.Printf("Частоты символов для Keylen=%d:\n", keylen)
			fmt.Println(freqs)

			for i := 0; i < keylen; i++ {
				minChi2 := math.MaxFloat64
				bestShift := 0

				for s := 0; s < S; s++ {
					currentChi2 := 0.0
					colLen := v_i[i]
					if colLen == 0 {
						continue
					}
					for j := 0; j < S; j++ {
						obs := float64(freqs[i][(j+s)%S])
						exp := float64(n) * theor[alphabetRunes[j]]

						if exp > 0 {
							currentChi2 += math.Pow(obs-exp, 2) / exp
						}
					}
					if currentChi2 < minChi2 {
						minChi2 = currentChi2
						bestShift = s
					}
				}
				key = append(key, bestShift)
			}
		}
	}

	var sortedChi2 []kv
	for k, v := range chi2res {
		sortedChi2 = append(sortedChi2, kv{k, v})
	}
	sort.Slice(sortedChi2, func(i, j int) bool {
		return sortedChi2[i].Value > sortedChi2[j].Value
	})
	fmt.Println("\nРезультаты хи-квадрат:")
	fmt.Printf("%v\n", sortedChi2)

	fmt.Println("\nНайденный ключ (сдвиги):")
	fmt.Println(key)

	var result strings.Builder
	for i, char := range textRunes {
		charIndex, _ := alphabetIndexMap[char]

		keyShift := key[i%keylen]
		decryptedIndex := (charIndex - keyShift + S) % S
		result.WriteRune(alphabetRunes[decryptedIndex])
	}

	fmt.Println("\nРезультат:")
	fmt.Println(result.String())
}
