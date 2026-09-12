package main

import "fmt"

func main() {
	var chic int
	fmt.Print("Введите число: ")
	fmt.Scan(&chic)
	if chic%2 == 0 {
		fmt.Println("Чётное число", chic)
	} else {
		fmt.Println("Нечётное число", chic)
	}

}
