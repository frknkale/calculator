// Package calculator contains arithmetic rules independent of HTTP or UI code.
package calculator

import (
	"errors"
	"math"
)

var (
	ErrOperation      = errors.New("Choose a supported operation.")
	ErrOperand        = errors.New("Provide the required finite numeric operands.")
	ErrDivisionByZero = errors.New("Cannot divide by zero.")
	ErrNegativeRoot   = errors.New("Cannot take the square root of a negative number.")
	ErrResult         = errors.New("The result is outside the supported real-number range.")
)

// Calculate applies the selected operation. Square root uses only a.
// Percentage means b percent of a. Power follows math.Pow, including 0^0 = 1.
func Calculate(operation string, a, b *float64) (float64, error) {
	binary := false
	switch operation {
	case "add", "subtract", "multiply", "divide", "power", "percentage":
		binary = true
	case "sqrt":
	default:
		return 0, ErrOperation
	}
	if !finite(a) || (binary && !finite(b)) || (!binary && b != nil) {
		return 0, ErrOperand
	}

	var result float64
	switch operation {
	case "add":
		result = *a + *b
	case "subtract":
		result = *a - *b
	case "multiply":
		result = *a * *b
	case "divide":
		if *b == 0 {
			return 0, ErrDivisionByZero
		}
		result = *a / *b
	case "power":
		result = math.Pow(*a, *b)
	case "sqrt":
		if *a < 0 {
			return 0, ErrNegativeRoot
		}
		result = math.Sqrt(*a)
	case "percentage":
		result = *a * (*b / 100)
	}
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return 0, ErrResult
	}
	return result, nil
}

func finite(value *float64) bool {
	return value != nil && !math.IsNaN(*value) && !math.IsInf(*value, 0)
}
