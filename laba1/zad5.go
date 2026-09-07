package main

import "fmt"

func calculate(a float64, b float64) (float64, float64) {
	sum := a + b
	diff := a - b
	return sum, diff
}

func main() {
	num1 := 13.6
	num2 := 5.6

	totalSum, totalDiff := calculate(num1, num2)
	fmt.Println(num1, num2)
	fmt.Println(totalSum, totalDiff)
}
