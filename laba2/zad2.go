package main

import (
	"fmt"
)

func checkSign(n int) string {
	switch {
	case n > 0:
		return "Positive"
	case n < 0:
		return "Negative"
	default:
		return "Zero"
	}
}

func main() {
	var num int
	fmt.Print("Введите число: ")
	fmt.Scan(&num)
	fmt.Println(checkSign(num))
}
