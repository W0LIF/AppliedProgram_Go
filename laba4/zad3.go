package main

import "fmt"

func main() {
	userAges := map[string]int{
		"Oleg":   23,
		"Miha":   20,
		"Nikita": 20,
	}

	delete(userAges, "Oleg")

	for name, age := range userAges {
		fmt.Println(name, age)
	}
}
