package main

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"laba9/models"
)

func readLine(in *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	line, _ := in.ReadString('\n')
	return strings.TrimSpace(line)
}

func readInt(in *bufio.Reader, prompt string) int {
	for {
		s := readLine(in, prompt)
		n, err := strconv.Atoi(s)
		if err == nil {
			return n
		}
		fmt.Println("нужно целое число, попробуй ещё раз")
	}
}

func reportError(err error) {
	if err == nil {
		return
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		fmt.Printf("Ошибка [%d]: %s\n", apiErr.Status, apiErr.Message)
		for field, msg := range apiErr.Details {
			fmt.Printf("   • %s: %s\n", field, msg)
		}
		return
	}
	fmt.Printf("%v\n", err)
}

func printUsers(users []models.User) {
	if len(users) == 0 {
		fmt.Println("(пользователей нет)")
		return
	}
	fmt.Printf("%-5s %-20s %-30s %-5s\n", "ID", "NAME", "EMAIL", "AGE")
	fmt.Println(strings.Repeat("─", 65))
	for _, u := range users {
		fmt.Printf("%-5d %-20s %-30s %-5d\n", u.ID, u.Name, u.Email, u.Age)
	}
}

func printUser(u models.User) {
	fmt.Printf("  ID:    %d\n", u.ID)
	fmt.Printf("  Name:  %s\n", u.Name)
	fmt.Printf("  Email: %s\n", u.Email)
	fmt.Printf("  Age:   %d\n", u.Age)
}
