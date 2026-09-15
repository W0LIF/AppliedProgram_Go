package main

import "fmt"

func main() {
	var numbers [5]int

	fmt.Println("Введите 5 целых чисел: ")
	for i := 0; i < 5; i++ {
		fmt.Scan(&numbers[i])
	}

	fmt.Println("В обратном порядке: ")
	for i := 4; i >= 0; i-- {
		fmt.Printf("%d", numbers[i])
	}
}
