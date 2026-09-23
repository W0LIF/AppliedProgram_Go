package main

import (
	"fmt"
	"sync"
)

type Request struct {
	a, b float64
	op   string
	resp chan float64
}

func calculatorServer(requests <-chan Request) {
	for req := range requests {
		var result float64
		switch req.op {
		case "+":
			result = req.a + req.b
		case "-":
			result = req.a - req.b
		case "*":
			result = req.a * req.b
		case "/":
			if req.b == 0 {
				fmt.Println("Ошибка деление на ноль")
				result = 0
			} else {
				result = req.a / req.b
			}
		default:
			fmt.Println("Неизвестная операция", req.op)
			result = 0
		}
		req.resp <- result
	}
}

func main() {
	requests := make(chan Request)
	go calculatorServer(requests)
	var wg sync.WaitGroup
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			a := float64(id)
			b := float64(id + 1)
			ops := []string{"+", "-", "*", "/"}
			for _, op := range ops {
				resp := make(chan float64)
				requests <- Request{a, b, op, resp}
				result := <-resp
				fmt.Printf("Клиент %d: %.0f %s %.0f = %.2f\n", id, a, op, b, result)
			}
		}(i)
	}
	wg.Wait()
	close(requests)
	fmt.Println("Калькулятор завершён")
}
