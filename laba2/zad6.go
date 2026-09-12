package main

import "fmt"

func average(a, b int) float64 {
	return float64(a+b) / 2
}

func main() {
	a, b := 4, 7
	fmt.Printf("Среднее значение %d и %d: %.2f\n", a, b, average(a, b))
}
