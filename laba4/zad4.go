package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Print("Введите строку: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Ошибка при чтении ввода:", err)
		return
	}

	input = strings.TrimSpace(input)

	upperCaseOutput := strings.ToUpper(input)

	fmt.Println("Результат:", upperCaseOutput)
}
