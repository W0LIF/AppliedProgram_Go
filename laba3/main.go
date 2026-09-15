package main

import (
	"bufio"
	"fmt"
	"laba3/mathutils"
	"laba3/stringutils"
	"os"
	"strings"
)

func main() {
	var n int
	fmt.Print("Введите число: ")
	fmt.Scan(&n)

	result, err := mathutils.Factorial(n)
	if err != nil {
		fmt.Println("Ошибка", err)
		return
	}

	fmt.Printf("Факториал %d = %d\n", n, result)

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
	fmt.Print("Введите строку: ")
	s, _ := reader.ReadString('\n')
	s = strings.TrimSpace(s)

	reversed := stringutils.Reverse(s)
	fmt.Println("Перевёрнутая строка:", reversed)
}
