package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
)

// substitutionMap — гипотеза о заменах: зашифрованный символ → открытый
var substitutionMap = map[rune]rune{
	'К': 'Т', 'Д': 'Ч', 'Т': 'К', 'Ф': 'З', 'Н': 'П',
	'Г': 'Ш', 'У': 'И', 'М': 'Р', 'С': 'Л', 'Е': 'Ц',
	'Ц': 'Е', 'О': 'О', 'Р': 'М', 'Ю': 'Ю', 'Ь': 'А',
	'П': 'Н', 'Щ': 'В', 'Ч': 'Д', 'А': 'Ь', 'З': 'Ф',
	'Ы': 'Б', 'Б': 'Ы', 'И': 'У', 'Х': 'Ж', 'Ш': 'Г',
	'Л': 'С', 'Э': 'Я', 'Ж': 'Х', 'В': 'Щ', 'Я': 'Э',
}

type charState struct {
	orig    rune // исходный символ
	current rune // текущее значение (после замен)
}

type pos struct {
	current  rune
	replaced bool
}

func main() {
	// Читаем исходный текст (сохраняя пробелы и разрывы строк)
	text, err := os.ReadFile("3В.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения файла: %v\n", err)
		os.Exit(1)
	}
	input := string(text)

	// Анализируем частоты
	fmt.Println(Cyan + "Анализ зашифрованного текста" + Reset)
	fmt.Println(strings.Repeat("-", 60))
	printFrequencyAnalysis(input, 10)
	fmt.Println()
	printTrigramAnalysis(input, 15)
	fmt.Println(strings.Repeat("=", 60))

	// Преобразуем текст по гипотезе замен
	result := applySubstitutions(input)

	// Выводим результат
	fmt.Println(Cyan + "Расшифрованный текст:" + Reset)
	fmt.Println(colorize(result, substitutionMap))
	fmt.Println(strings.Repeat("-", 60))
}

// applySubstitutions — применяет замены по мапе, сохраняя структуру текста
func applySubstitutions(s string) []charState {
	var result []charState
	for _, r := range s {
		if to, ok := substitutionMap[r]; ok {
			result = append(result, charState{orig: r, current: to})
		} else {
			result = append(result, charState{orig: r, current: r})
		}
	}
	return result
}

// colorize — возвращает строку с подсветкой заменённых символов красным
func colorize(states []charState, subMap map[rune]rune) string {
	var sb strings.Builder
	for _, cs := range states {
		if cs.orig != cs.current {
			sb.WriteString(Red)
			sb.WriteRune(cs.current)
			sb.WriteString(Reset)
		} else {
			sb.WriteRune(cs.current)
		}
	}
	return sb.String()
}

// printFrequencyAnalysis — выводит топ N самых частых символов
func printFrequencyAnalysis(text string, n int) {
	freq := make(map[rune]int)
	for _, r := range text {
		if r != ' ' && r != '\n' && r != '\r' && r != '\t' {
			freq[r]++
		}
	}

	type kv struct {
		Key   rune
		Value int
	}
	var sorted []kv
	for k, v := range freq {
		sorted = append(sorted, kv{k, v})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	fmt.Println(Yellow + "Частоты символов:" + Reset)
	for i := 0; i < n && i < len(sorted); i++ {
		fmt.Printf("  '%c': %d\n", sorted[i].Key, sorted[i].Value)
	}
}

// printTrigramAnalysis — выводит топ N самых частых 3-грамм
func printTrigramAnalysis(text string, n int) {
	// Очищаем от пробелов и разрывов для поиска слов
	clean := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			return -1
		}
		return r
	}, text)

	trigrams := make(map[string]int)
	for i := 0; i <= len([]rune(clean))-3; i++ {
		trigram := string([]rune(clean)[i : i+3])
		trigrams[trigram]++
	}

	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	for k, v := range trigrams {
		sorted = append(sorted, kv{k, v})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	fmt.Println(Yellow + "Частые 3-граммы:" + Reset)
	for i := 0; i < n && i < len(sorted); i++ {
		fmt.Printf("  %s: %d\n", sorted[i].Key, sorted[i].Value)
	}
}



func readAll(path string) string {
	f, _ := os.Open(path)
	defer f.Close()

	var sb strings.Builder
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		sb.WriteString(sc.Text())
	}
	return sb.String()
}

func stringOfPositions(ps []pos, highlight bool) string {
	var sb strings.Builder
	for _, p := range ps {
		ch := p.current
		if highlight && p.replaced {
			sb.WriteString(Red)
			sb.WriteRune(ch)
			sb.WriteString(Reset)
		} else {
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}