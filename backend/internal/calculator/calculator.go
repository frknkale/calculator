// Package calculator contains arithmetic rules independent of HTTP or UI code.
package calculator

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/shopspring/decimal"
)

const significantDigits int32 = 16

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
		exponent := order(x) - order(y)
		if x.Abs().LessThan(y.Abs().Shift(exponent)) {
			exponent--
		}
		result = x.DivRound(y, significantDigits-1-exponent)
	case "power":
		if y.Abs().GreaterThan(maxPower) {
			return "", ErrPowerLimit
		}
		if x.IsZero() && y.IsZero() {
			result = decimal.NewFromInt(1)
		} else {
			result, err = x.PowWithPrecision(y, significantDigits)
			// Estimate the result before choosing decimal places.
			if err == nil && !result.IsZero() && !(y.IsInteger() && !y.IsNegative()) && order(result) >= -significantDigits {
				places := significantDigits - 1 - order(result)
				result, err = x.PowWithPrecision(y, significantDigits-1-order(result))
				if err == nil && significantDigits-1-order(result) > places {
					result, err = x.PowWithPrecision(y, significantDigits-1-order(result))
				}
			}
		}
	case "sqrt":
		if x.IsNegative() {
			return "", ErrNegativeRoot
		}
		exponent := order(x) / 2
		if order(x) < 0 && order(x)%2 != 0 {
			exponent--
		}
		result, err = x.SqrtRound(significantDigits - 1 - exponent)
	case "percentage":
		result = x.Mul(y.Shift(-2))
	}
	if err != nil {
		return "", ErrResult
	}
	return formatResult(result), nil
}

// order is the exponent in a value's normalized scientific notation.
func order(value decimal.Decimal) int32 {
	return int32(value.NumDigits()) + value.Exponent() - 1
}

func formatResult(value decimal.Decimal) string {
	if value.IsZero() {
		return "0"
	}
	value = value.Round(significantDigits - 1 - order(value))
	exponent := order(value)
	if exponent >= significantDigits || exponent < -6 {
		suffix := strconv.Itoa(int(exponent))
		if exponent >= 0 {
			suffix = "+" + suffix
		}
		return value.Shift(-exponent).String() + "e" + suffix
	}
	return value.String()
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
