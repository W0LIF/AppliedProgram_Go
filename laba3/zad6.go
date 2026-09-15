package main

import (
	"fmt"
	"unicode/utf8"
)

func findLongestString(slice []string) string {
	if len(slice) == 0 {
		return "пустой срез"
	}

	longest := slice[0]
	maxLen := utf8.RuneCountInString(longest)

	for _, str := range slice[1:] {
		currentLen := utf8.RuneCountInString(str)
		if currentLen > maxLen {
			maxLen = currentLen
			longest = str
		}
	}

	return longest
}

func main() {
	words := []string{"Go", "Язык программирования", "Срез", "Строка"}
	longestWord := findLongestString(words)

	if longestWord != "" {
		fmt.Printf("Срез строки: %v\n", words)
		fmt.Printf("Самая длинная строка среза: \"%s\" (длина: %d симв.)\n", longestWord, utf8.RuneCountInString(longestWord))
	} else {
		fmt.Println("Срез пуст")
	}
}
