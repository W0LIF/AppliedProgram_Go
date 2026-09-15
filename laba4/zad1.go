package main

import "fmt"

func averageUserAge(ages map[string]int) float64 {
	if len(ages) == 0 {
		return 0
	}

	totalAge := 0
	for _, age := range ages {
		totalAge += age
	}

	return float64(totalAge) / float64(len(ages))
}

func main() {
	userAges := map[string]int{
		"Oleg": 23,
		"Miha": 20,
	}

	userAges["Nikita"] = 20

	for name, age := range userAges {
		fmt.Println(name, age)
	}

	average := averageUserAge(userAges)
	fmt.Printf("\nСредний возраст: %.2f\n", average)
}
