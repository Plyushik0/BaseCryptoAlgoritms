package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
)

const ALPHABET = "абвгдежзийклмнопрстуфхцчшщъыьэюя"



func main() {

	files := []string{"etalon_russian.txt", "normal_russian.txt", "random_russian.txt"}
	names := []string{"etalon", "meaningful", "random"}

	uniFreq := make(map[string]map[string]int)
	biFreq := make(map[string]map[string]int)
	triFreq := make(map[string]map[string]int)
	uniChi := make(map[string]float64)
	biChi := make(map[string]float64)
	triChi := make(map[string]float64)

	var etalon_letters map[rune]int
	var etalon_pairs map[string]int
	var etalon_trigrams map[string]int
	var etalon_total_letters int
	var etalon_total_pairs int
	var etalon_total_trigrams int

	for i, file := range files {
		text, _ := os.ReadFile(file)
		
		clean := cleanText(string(text))

		name := names[i]

		letters := countUnigrams(clean)
		uniFreq[name] = toStringMap(letters)

		pairs := countBigrams(clean)
		biFreq[name] = pairs

		trigrams := countTrigrams(clean)
		triFreq[name] = trigrams

		
		if name == "etalon" {
			etalon_letters = letters
			etalon_pairs = pairs
			etalon_trigrams = trigrams
			etalon_total_letters = len([]rune(clean))
			etalon_total_pairs = len([]rune(clean)) - 1
			etalon_total_trigrams = len([]rune(clean)) - 2
		} else {
			
			uniChi[name] = chi2Unigams(letters, etalon_letters, etalon_total_letters)
			biChi[name] = chi2Bigrams(pairs, etalon_pairs, etalon_total_pairs)
			triChi[name] = chi2Trigrams(trigrams, etalon_trigrams, etalon_total_trigrams)
		}
	}


	save("unigram_frequencies.json", uniFreq)
	save("bigram_frequencies.json", biFreq)
	save("trigram_frequencies.json", triFreq)
	save("unigram_chi2_results.json", uniChi)
	save("bigram_chi2_results.json", biChi)
	save("trigram_chi2_results.json", triChi)


	fmt.Println("Осмысленный текст:")
	fmt.Printf("  Xi2 (биграммы)  = %.2f\n", biChi["meaningful"])
	fmt.Printf("  Xi2 (триграммы) = %.2f\n", triChi["meaningful"])

	fmt.Println("Рандомный текст:")
	fmt.Printf("  Xi2 (биграммы)  = %.2f\n", biChi["random"])
	fmt.Printf("  Xi2 (триграммы) = %.2f\n", triChi["random"])

	
}


func cleanText(s string) string {
	s = strings.ToLower(s)
	var result strings.Builder
	for _, r := range s {
		if strings.ContainsRune(ALPHABET, r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}


func countUnigrams(text string) map[rune]int {
	count := make(map[rune]int)
	for _, letter := range text {
		count[letter]++
	}
	return count
}


func countBigrams(text string) map[string]int {
	count := make(map[string]int)
	runes := []rune(text)
	for i := 0; i < len(runes)-1; i++ {
		pair := string(runes[i]) + string(runes[i+1])
		count[pair]++
	}
	return count
}


func countTrigrams(text string) map[string]int {
	count := make(map[string]int)
	runes := []rune(text)
	for i := 0; i < len(runes)-2; i++ {
		trigram := string(runes[i]) + string(runes[i+1]) + string(runes[i+2])
		count[trigram]++
	}
	return count
}


func chi2Unigams(observed map[rune]int, etalon map[rune]int, etalon_total int) float64 {
	var chi float64
	observed_total := 0
	for _, count := range observed {
		observed_total += count
	}

	for _, letter := range ALPHABET {
		obs := float64(observed[letter])
		p := float64(etalon[letter]) / float64(etalon_total)
		expected := p * float64(observed_total)
        chi += math.Pow(obs-expected, 2) / expected
		
	}
	return chi
}

func chi2Bigrams(observed map[string]int, etalon map[string]int, etalon_total int) float64 {
	var chi float64
	observed_total := 0
	for _, count := range observed {
		observed_total += count
	}

	for bigram, etalon_count := range etalon {
		obs := float64(observed[bigram])
		p := float64(etalon_count) / float64(etalon_total)
		expected := p * float64(observed_total)
		chi += math.Pow(obs-expected, 2) / expected

	}
	return chi
}

func chi2Trigrams(observed map[string]int, etalon map[string]int, etalon_total int) float64 {
	var chi float64
	observed_total := 0
	for _, count := range observed {
		observed_total += count
	}

	for trigram, etalon_count := range etalon {
		obs := float64(observed[trigram])
		p := float64(etalon_count) / float64(etalon_total)
		expected := p * float64(observed_total)
		chi += math.Pow(obs-expected, 2) / expected
	}
	return chi
}



func toStringMap(m map[rune]int) map[string]int {
	res := make(map[string]int)
	for r, c := range m {
		res[string(r)] = c
	}
	return res
}

func save(filename string, data any) {
	file, _ := os.Create(filename)
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
}