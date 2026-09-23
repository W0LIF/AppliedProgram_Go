package main

import (
	"fmt"
	"math/rand"
	"time"
)

func calculcateFactrorial() {
	time.Sleep(2 * time.Second)
	result := 1
	for i := 2; i <= 5; i++ {
		result *= i
	}
	fmt.Println("Факториал 5 равен:", result)
}

func generateRandom() {
	time.Sleep(1 * time.Second)
	num := rand.Intn(100)
	fmt.Println("Случайное число:", num)
}

func calculateSum() {
	time.Sleep(1 * time.Second)
	sum := 0
	for i := 1; i <= 10; i++ {
		sum += i
	}
	fmt.Println("Сумма чисел от 1 до 10:", sum)
}

func main() {
	fmt.Println("Запускаем горутины 3 задач")
	go calculcateFactrorial()
	go generateRandom()
	go calculateSum()
	time.Sleep(3 * time.Second)
	fmt.Println("Программа завершена")
}
