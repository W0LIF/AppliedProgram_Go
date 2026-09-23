package main

import (
	"fmt"
	"sync"
)

func main() {
	const goroutine = 1000
	const increment = 1000

	var counter int
	var wg sync.WaitGroup
	wg.Add(goroutine)
	for i := 0; i < goroutine; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < increment; j++ {
				counter++
			}
		}()
	}
	wg.Wait()
	fmt.Println("Без мьютекса:", counter)

	counter = 0
	var mu sync.Mutex
	wg.Add(goroutine)
	for i := 0; i < goroutine; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < increment; j++ {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	fmt.Println("С мьютексом:", counter)
}
