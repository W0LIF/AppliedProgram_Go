package main

import "fmt"

type Stringer interface {
	String() string
}

type Book struct {
	Title  string
	Author string
	Year   int
}

func (b Book) String() string {
	return fmt.Sprintf("Книга: %s (%s, %d)", b.Title, b.Author, b.Year)
}

func PrintCustomString(s Stringer) {
	fmt.Println(s.String())
}

func main() {
	myBook := Book{
		Title:  "Ураганные войны",
		Author: "Теа Гуанзон",
		Year:   2024,
	}

	PrintCustomString(myBook)
}
