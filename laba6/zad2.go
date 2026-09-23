package main

import (
	"fmt"
	"time"
)

func generateFibonacci(ch chan int) {
	a, b := 0, 1
	for i := 0; i < 10; i++ {
		ch <- a
		a, b = b, a+b
		time.Sleep(300 * time.Millisecond)
	}
	close(ch)
}

func main() {
	ch := make(chan int)
	go generateFibonacci(ch)
	for num := range ch {
		fmt.Printf("Получено число Фибоначи: %d\n", num)
	}

	fmt.Println("Программа завершена")
}
