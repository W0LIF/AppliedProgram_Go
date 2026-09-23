package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func worker(id int, jobs <-chan string, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	for line := range jobs {
		reversed := reverse(line)
		results <- fmt.Sprintf("Воркер %d: %s -> %s", id, line, reversed)
	}
}

func main() {
	var workers int
	fmt.Print("Введите количество воркеров: ")
	fmt.Scan(&workers)
	if workers < 1 {
		workers = 1
	}

	inputFile := "input.txt"
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		os.WriteFile(inputFile, []byte("hello\nworld\nGo programming\n"), 0644)
	}

	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Println("Ошибка открытия файла:", err)
		return
	}
	defer file.Close()

	jobs := make(chan string)
	results := make(chan string)

	var wg sync.WaitGroup
	for i := 1; i <= workers; i++ {
		wg.Add(1)
		go worker(i, jobs, results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	go func() {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			jobs <- scanner.Text()
		}
		close(jobs)
	}()

	outputFile := "output.txt"
	out, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Ошибка создания файла:", err)
		return
	}
	defer out.Close()

	for res := range results {
		fmt.Println(res)
		fmt.Fprintln(out, res)
	}

	fmt.Println("Готово. Результаты сохранены в", outputFile)
}
