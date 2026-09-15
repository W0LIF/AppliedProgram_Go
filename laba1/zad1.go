package main

import (
	"fmt"
	"time"
)

func main() {
	currentTime := time.Now()
	fmt.Println("Текущее время", currentTime.Format("02.01.2006 15:04:05"))
}
