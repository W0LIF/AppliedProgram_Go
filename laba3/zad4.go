package main

import (
	"fmt"
	"slices"
)

func main() {
	var nums [5]int

	for i := 0; i < len(nums); i++ {
		nums[i] = (i + 1)
	}

	fmt.Println("Весь массив:", nums)

	numis := nums[1:4]

	fmt.Println("Срез массива", numis)

	numis = append(numis, 3)

	fmt.Println("Добавление элемента", numis)

	numis = slices.Delete(numis, 2, 3)

	fmt.Println("Удаление числа из массива", numis)
}
