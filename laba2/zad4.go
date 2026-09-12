package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

func stringLength(s string) int {
	return utf8.RuneCountInString(s)
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Введите строку: ")
	s, _ := reader.ReadString('\n')
	s = strings.TrimSpace(s)

	fmt.Println("Длина строки:", stringLength(s))
}
