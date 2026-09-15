package main

import "fmt"

type Person struct {
	name string
	age  int
}

func (p *Person) PrintInfo() {
	fmt.Printf("Имя %s, Возраст: %d\n", p.name, p.age)
}

func (p *Person) Birthday() {
	p.age++
}

func main() {
	person := Person{
		name: "Oleg",
		age:  23,
	}

	fmt.Println("Задание 1")
	person.PrintInfo()
	person.Birthday()
	fmt.Println("Задание 2")
	person.PrintInfo()
}
