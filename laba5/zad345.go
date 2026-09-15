package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
}

type Circle struct {
	radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * math.Pow(c.radius, 2)
}

type Rectangle struct {
	width, height float64
}

func (r Rectangle) Area() float64 {
	return r.width * r.height
}

func printShapesAreas(shapes []Shape) {
	for i, shape := range shapes {
		fmt.Printf("Фигура №%d: площадь = %.2f\n", i+1, shape.Area())
	}
}

func main() {
	c1 := Circle{radius: 3.0}
	c2 := Circle{radius: 4.5}
	r1 := Rectangle{width: 3.0, height: 4.5}
	r2 := Rectangle{width: 5.0, height: 6.0}
	shapesList := []Shape{c1, c2, r1, r2}
	fmt.Println("Вывод всех площадей фигур")
	printShapesAreas(shapesList)

}
