// Package calculator contains arithmetic rules independent of HTTP or UI code.
package calculator

import (
	"errors"
	"regexp"

	"github.com/shopspring/decimal"
)

const decimalPlaces = 16

var (
	ErrOperation      = errors.New("Choose a supported operation.")
	ErrOperand        = errors.New("Provide the required decimal strings, up to 128 characters with decimal exponents between -1000 and 1000.")
	ErrDivisionByZero = errors.New("Cannot divide by zero.")
	ErrNegativeRoot   = errors.New("Cannot take the square root of a negative number.")
	ErrPowerLimit     = errors.New("Power exponent must be between -1000 and 1000.")
	ErrResult         = errors.New("The result is not a supported real number.")
	decimalSyntax     = regexp.MustCompile(`^[+-]?(?:\d+\.?\d*|\.\d+)(?:[eE][+-]?\d+)?$`)
	maxPower          = decimal.NewFromInt(1000)
)

// Calculate accepts and returns decimal strings to preserve their digits.
// Percentage means b percent of a. Power defines 0^0 as 1.
func Calculate(operation string, a, b *string) (string, error) {
	switch operation {
	case "add", "subtract", "multiply", "divide", "power", "percentage":
		if b == nil {
			return "", ErrOperand
		}
	case "sqrt":
		if b != nil {
			return "", ErrOperand
		}
	default:
		return "", ErrOperation
	}
	x, err := parseOperand(a)
	if err != nil {
		return "", err
	}
	var y decimal.Decimal
	if b != nil {
		y, err = parseOperand(b)
		if err != nil {
			return "", err
		}
	}

	var result decimal.Decimal
	switch operation {
	case "add":
		result = x.Add(y)
	case "subtract":
		result = x.Sub(y)
	case "multiply":
		result = x.Mul(y)
	case "divide":
		if y.IsZero() {
			return "", ErrDivisionByZero
		}
		// Use extra decimal places to preserve small quotients.
		places := int64(decimalPlaces) + max(0, order(y)-order(x))
		result = x.DivRound(y, int32(places))
	case "power":
		if y.Abs().GreaterThan(maxPower) {
			return "", ErrPowerLimit
		}
		if x.IsZero() && y.IsZero() {
			result = decimal.NewFromInt(1)
		} else {
			result, err = x.PowWithPrecision(y, decimalPlaces)
		}
	case "sqrt":
		if x.IsNegative() {
			return "", ErrNegativeRoot
		}
		// Use extra decimal places to preserve small square roots.
		places := decimalPlaces + max(0, -x.Exponent()/2)
		result, err = x.SqrtRound(places)
	case "percentage":
		result = x.Mul(y.Shift(-2))
	}
	if err != nil {
		return "", ErrResult
	}
	return result.String(), nil
}

// order is the exponent in a value's normalized scientific notation.
func order(value decimal.Decimal) int64 {
	return int64(value.NumDigits()) + int64(value.Exponent()) - 1
}

func parseOperand(value *string) (decimal.Decimal, error) {
	if value == nil || len(*value) > 128 || !decimalSyntax.MatchString(*value) {
		return decimal.Decimal{}, ErrOperand
	}
	number, err := decimal.NewFromString(*value)
	if err != nil || number.Exponent() < -1000 || number.Exponent() > 1000 {
		return decimal.Decimal{}, ErrOperand
	}
	return number, nil
}
