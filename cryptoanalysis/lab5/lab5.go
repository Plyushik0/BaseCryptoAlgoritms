package main

import (
	"fmt"
	"os"
)

func main() {
	var filename string
	var choice int
	var gamma string

	fmt.Print("имя файла: ")
	fmt.Scanln(&filename)

	C, _ := os.ReadFile(filename)

	fmt.Println("Выберите действие:")
	fmt.Println("1 - Зашифровать")
	fmt.Println("2 - Расшифровать")
	fmt.Scanln(&choice)

	fmt.Print("Гамма: ")
	fmt.Scanln(&gamma)

	g := []byte(gamma)

	blockSize := len(g)
	for i := 0; i < len(C); i++ {
		M := C[i]
		pos := i % blockSize
		Y := g[pos]

		if choice == 1 {
			C[i] = byte((int(M) + int(Y)) % 256)
		} else if choice == 2 {
			C[i] = byte((int(M) - int(Y) + 256) % 256)
		}
	}

	output := "out.txt"
	if choice == 2 {
		output = "out1.txt"
	}

	os.WriteFile(output, C, 0644)

}
