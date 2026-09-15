package mathutils

import "fmt"

func Factorial(n int) (uint64, error) {
	if n < 0 {
		return 0, fmt.Errorf("факториал не определён для отрицательных чисел")
	}
	if n > 20 {
		return 0, fmt.Errorf("слишком большое число: переполнение uint64")
	}

	var result uint64 = 1
	for i := 2; i <= n; i++ {
		result *= uint64(i)
	}
	return result, nil
}
