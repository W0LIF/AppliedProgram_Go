package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	randomCh := make(chan int)
	parityInput := make(chan int)
	parityCh := make(chan string)

	go func() {
		for i := 0; i < 10; i++ {
			num := rand.Intn(100)
			randomCh <- num
			parityInput <- num
			time.Sleep(200 * time.Millisecond)
		}
		close(randomCh)
		close(parityInput)
	}()

	go func() {
		for num := range parityInput {
			if num%2 == 0 {
				parityCh <- fmt.Sprintf("Число %d чётное", num)
			} else {
				parityCh <- fmt.Sprintf("Число %d нечётное", num)
			}
		}
		close(parityCh)
	}()

	for randomCh != nil || parityCh != nil {
		select {
		case num, ok := <-randomCh:
			if !ok {
				randomCh = nil
				continue
			}
			fmt.Println("Случайное число:", num)
		case msg, ok := <-parityCh:
			if !ok {
				randomCh = nil
				continue
			}
			fmt.Println("Чётность:", msg)
		}
	}
	fmt.Println("Программа завершена")
}
