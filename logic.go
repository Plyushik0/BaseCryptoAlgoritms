package main

import (
	"math/rand"
	"os"
	"strconv"
	"strings"
	"bufio"
	"fmt"


)
func manualInput(name string) []float64 {
	scanner := bufio.NewScanner(os.Stdin)
	var arr []float64

	fmt.Printf("Введите значения массива %s через пробел\n", name)
	scanner.Scan()
	input := scanner.Text()

	for _, s := range strings.Fields(input) {
		val, _ := strconv.ParseFloat(s, 64)
		arr = append(arr, val)
	}

	return arr
}

func generateArray(size int) []float64 {
	r := rand.New(rand.NewSource(rand.Int63())) 
	array := make([]float64, size)
	for i := 0; i < size; i++ {
		array[i] = float64(r.Intn(10)) / 10.0 
	}
	return array
}

func AND(x1, x2 float64) float64 {
	if x1 < x2 {
		return x1
	}
	return x2
}

func OR(x1, x2 float64) float64 {
	if x1 > x2 {
		return x1
	}
	return x2
}

func NOT(x1 float64) float64 {
	return 1 - x1
}

func IMPL(x1, x2 float64) float64 {
	return OR(NOT(x1), x2)
}

func EQUIV(x1, x2 float64) float64 {
	return AND(x1, x2)
}

func F1(x1, x2 float64) float64 {
	return IMPL(NOT(x1), x2)
}

func F2(x1, x2 float64) float64 {
	return AND(x1, NOT(x2))
}

func COMPUTE(X1, X2 []float64) float64{
	if len(X1) == 0 || len(X2) == 0 {
		panic("Один из массивов пуст!")
	}

	minResult := 1.0 
	for _, x1 := range X1 {
		for _, x2 := range X2 {
			current := EQUIV(F1(x1, x2), F2(x1, x2))
			if current < minResult {
				minResult = current
			}
		}
	}
	return minResult
}
