package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	RedColor   = "\033[31m"
	GreenColor = "\033[32m"
	ResetColor = "\033[0m"
)

func main() {

	cipherText := readFile("cipher.txt")
	lines := strings.Split(strings.TrimSpace(cipherText), "\n")
	keyLen := len([]rune(lines[0]))
	breakCipher(lines, keyLen)
}


func breakCipher(lines []string, keyLen int) {
	
	columns := make([][]rune, keyLen)
	for _, line := range lines {
		runes := []rune(line)
		for col := 0; col < keyLen && col < len(runes); col++ {
			columns[col] = append(columns[col], runes[col])
		}
	}


	forbiddenTable := buildForbiddenTable(columns, keyLen)

	fmt.Println("Таблица запретных биграмм:")
	printForbiddenTable(forbiddenTable, keyLen)
	printTransition(forbiddenTable, keyLen)
	fmt.Println()

	
	fmt.Println(GreenColor + "Ключ" + ResetColor)
	key := KeyOrder(forbiddenTable, keyLen)
	fmt.Printf("Порядок столбцов: %v\n\n", key)


	decrypted := decryptWithKey(columns, key)
	decryptedClean := strings.ReplaceAll(decrypted, "\n", "")
	
	fmt.Println(GreenColor + "РЕЗУЛЬТАТ" + ResetColor)
	fmt.Println("Первые 400 символов:")
	if len(decryptedClean) > 400 {
		fmt.Println(decryptedClean[:400])
	} else {
		fmt.Println(decryptedClean)
	}
	fmt.Println()
	
	
	fmt.Println(GreenColor + "DEBUG" + ResetColor)
	showDecryptionProcess(columns, key)
	fmt.Println()
	
	saveToFile("decrypted.txt", decryptedClean)
	
}


// *Ь не может идти после гласных/пробела, *Й не может идти после согласных/пробела
func buildForbiddenTable(columns [][]rune, keyLen int) [][]bool {
	table := make([][]bool, keyLen)
	for i := range table {
		table[i] = make([]bool, keyLen)
	}

	for i := 0; i < keyLen; i++ {
		for j := 0; j < keyLen; j++ {
			if i == j {
				continue
			}

			forbidden := false
			minLen := len(columns[i])
			if len(columns[j]) < minLen {
				minLen = len(columns[j])
			}

			for k := 0; k < minLen; k++ {
				ch1 := columns[i][k]
				ch2 := columns[j][k]

				if (ch2 == 'Ь') && (isVowel(ch1) || ch1 == '_') {
					forbidden = true
					break
				}
				if (ch2 == 'Й') && (isConsonant(ch1) || ch1 == '_') {
					forbidden = true
					break
				}
			}

			table[i][j] = forbidden
		}
	}

	return table
}


func isVowel(ch rune) bool {
	vowels := []rune("АЕИОУЮЯЁ")
	for _, v := range vowels {
		if v == ch {
			return true
		}
	}
	return false
}


func isConsonant(ch rune) bool {
	consonants := []rune("БВГДЖЗЙКЛМНПРСТФХЦЧШЩ")
	for _, c := range consonants {
		if c == ch {
			return true
		}
	}
	return false
}


func printForbiddenTable(table [][]bool, keyLen int) {
	const (
		Red    = "\033[31m"
		Green  = "\033[32m"
		Reset  = "\033[0m"
	)

	fmt.Print("    ")
	for j := 0; j < keyLen; j++ {
		fmt.Printf("%3d ", j)
	}
	fmt.Println()

	fmt.Print("   ")
	for j := 0; j < keyLen; j++ {
		fmt.Print("────")
	}
	fmt.Println()

	for i := 0; i < keyLen; i++ {
		fmt.Printf(" %2d:", i)
		for j := 0; j < keyLen; j++ {
			if i == j {
				fmt.Print("  - ")
			} else if table[i][j] {
				fmt.Printf(" %s X %s", Red, Reset)
			} else {
				fmt.Printf(" %s O %s", Green, Reset)
			}
		}
		fmt.Println()
	}
	fmt.Println()
}
func printTransition(table [][]bool, keyLen int) {
	const (
		Red    = "\033[31m"
		Green  = "\033[32m"
		Yellow = "\033[33m"
		Reset  = "\033[0m"
	)

	fmt.Println(Yellow + "Переходы" + Reset)
	for i := 0; i < keyLen; i++ {
		allowed := []int{}
		for j := 0; j < keyLen; j++ {
			if i != j && !table[i][j] {
				allowed = append(allowed, j)
			}
		}

		fmt.Printf(" %2d -> ", i)
		if len(allowed) == 0 {
			fmt.Printf("%sнет переходов%s\n", Red, Reset)
		} else if len(allowed) == 1 {
			fmt.Printf("%s%d (однозначно)%s\n", Green, allowed[0], Reset)
		} else {
			
			for idx, j := range allowed {
				if idx > 0 {
					fmt.Print(" ")
				}
				fmt.Printf("%s%d%s", Green, j, Reset)
			}
			fmt.Println()
		}
	}
	fmt.Println()
}
func KeyOrder(table [][]bool, keyLen int) []int {
	order := make([]int, 0, keyLen)
	used := make([]bool, keyLen)

	pred := make([]int, keyLen)
	for i := 0; i < keyLen; i++ {
		for j := 0; j < keyLen; j++ {
			if i != j && !table[j][i] { 
				pred[i]++
			}
		}
	}

	for len(order) < keyLen {
		starts := []int{}
		for i := 0; i < keyLen; i++ {
			if !used[i] && pred[i] == 0 {
				starts = append(starts, i)
			}
		}

		if len(starts) == 0 {

			for i := 0; i < keyLen; i++ {
				if !used[i] {
					order = append(order, i)
					used[i] = true
				}
			}
			break
		}

		start := starts[0]
		chain := buildChain(start, table, keyLen, used)

		for _, col := range chain {
			if !used[col] {
				order = append(order, col)
				used[col] = true

				for next := 0; next < keyLen; next++ {
					if !table[col][next] {
						pred[next]--
					}
				}
			}
		}
	}

	return order
}

func buildChain(start int, table [][]bool, keyLen int, used []bool) []int {
	chain := []int{start}
	current := start
	visited := make([]bool, keyLen)
	visited[start] = true

	for {
		nextCandidates := []int{}
		for j := 0; j < keyLen; j++ {
			if !visited[j] && !used[j] && !table[current][j] {
				nextCandidates = append(nextCandidates, j)
			}
		}

		if len(nextCandidates) != 1 {
			break
		}

		next := nextCandidates[0]
		chain = append(chain, next)
		visited[next] = true
		current = next
	}

	return chain
}


func decryptWithKey(columns [][]rune, keyOrder []int) string {

	colLen := len(columns[keyOrder[0]])
	var result strings.Builder
	for row := 0; row < colLen; row++ {
		for _, colIdx := range keyOrder {
			if colIdx < len(columns) && row < len(columns[colIdx]) {
				result.WriteRune(columns[colIdx][row])
			}
		}
		result.WriteRune('\n')
	}

	return result.String()
}

func readFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("Ошибка: не могу открыть файл %s\n", path)
		return ""
	}
	defer f.Close()

	var sb strings.Builder
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		sb.WriteString(scanner.Text())
		sb.WriteRune('\n')
	}
	return sb.String()
}
func saveToFile(path string, content string) error {
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("Ошибка: не могу создать файл %s\n", path)
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return err
}


func showDecryptionProcess(columns [][]rune, keyOrder []int) {
	fmt.Println("Порядок столбцов:", keyOrder)
	fmt.Println("\nПервые 5 строк расшифровки:\n")
	
	rowsToShow := 5
	if len(columns[0]) < rowsToShow {
		rowsToShow = len(columns[0])
	}
	
	for row := 0; row < rowsToShow; row++ {
		fmt.Printf("Строка %d: ", row)
		for i, colIdx := range keyOrder {
			if i < 5 {  
				if row < len(columns[colIdx]) {
					fmt.Printf("[%d]='%c' ", colIdx, columns[colIdx][row])
				}
			}
		}
		fmt.Printf(" ... ")
		

		for _, colIdx := range keyOrder {
			if row < len(columns[colIdx]) {
				fmt.Printf("%c", columns[colIdx][row])
			}
		}
		fmt.Println()
	}
}
