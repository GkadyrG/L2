package main

import (
	"fmt"
	"sort"
	"strings"
)

func main() {
	words := []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}
	anagrams := findAnagrams(words)

	for k, v := range anagrams {
		fmt.Printf("%q: %v\n", k, v)
	}
}

func findAnagrams(words []string) map[string][]string {
	temp := make(map[string][]string)
	for _, word := range words {
		key := sortStr(word)
		temp[key] = append(temp[key], strings.ToLower(word))
	}

	result := make(map[string][]string)
	for _, group := range temp {
		if len(group) > 1 {
			sort.Strings(group)
			result[group[0]] = group
		}
	}

	return result

}

func sortStr(word string) string {
	wrd := strings.ToLower(word)
	letters := strings.Split(wrd, "")
	sort.Strings(letters)
	sortedWord := strings.Join(letters, "")
	return sortedWord
}
