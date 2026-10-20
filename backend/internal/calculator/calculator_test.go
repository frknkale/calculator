package calculator

import (
	"errors"
	"math"
	"testing"
)

func number(value float64) *float64 { return &value }

func TestCalculate(t *testing.T) {
	tests := []struct {
		name, operation string
		a, b            *float64
		want            float64
		wantErr         error
	}{
		{"addition", "add", number(2), number(3), 5, nil},
		{"negative decimal", "add", number(-1.5), number(0.25), -1.25, nil},
		{"subtraction order", "subtract", number(2), number(5), -3, nil},
		{"multiplication", "multiply", number(-4), number(2.5), -10, nil},
		{"zero is present", "multiply", number(7), number(0), 0, nil},
		{"division", "divide", number(7), number(2), 3.5, nil},
		{"zero numerator", "divide", number(0), number(2), 0, nil},
		{"division by zero", "divide", number(3), number(0), 0, ErrDivisionByZero},
		{"division by negative zero", "divide", number(3), number(math.Copysign(0, -1)), 0, ErrDivisionByZero},
		{"power", "power", number(2), number(10), 1024, nil},
		{"negative exponent", "power", number(2), number(-2), 0.25, nil},
		{"zero to zero", "power", number(0), number(0), 1, nil},
		{"imaginary power", "power", number(-1), number(0.5), 0, ErrResult},
		{"zero negative power", "power", number(0), number(-1), 0, ErrResult},
		{"square root", "sqrt", number(81), nil, 9, nil},
		{"zero root", "sqrt", number(0), nil, 0, nil},
		{"negative root", "sqrt", number(-1), nil, 0, ErrNegativeRoot},
		{"percentage", "percentage", number(200), number(25), 50, nil},
		{"negative percentage", "percentage", number(200), number(-50), -100, nil},
		{"negative percentage base", "percentage", number(-200), number(25), -50, nil},
		{"zero percent", "percentage", number(200), number(0), 0, nil},
		{"zero percentage base", "percentage", number(0), number(25), 0, nil},
		{"over one hundred percent", "percentage", number(200), number(150), 300, nil},
		{"large percentage base", "percentage", number(1e308), number(50), 5e307, nil},
		{"missing percentage rate", "percentage", number(200), nil, 0, ErrOperand},
		{"unknown operation", "modulo", number(1), number(2), 0, ErrOperation},
		{"missing operation", "", number(1), number(2), 0, ErrOperation},
		{"missing a", "add", nil, number(2), 0, ErrOperand},
		{"missing b", "add", number(1), nil, 0, ErrOperand},
		{"unexpected b", "sqrt", number(4), number(2), 0, ErrOperand},
		{"infinite a", "add", number(math.Inf(1)), number(2), 0, ErrOperand},
		{"nan b", "add", number(1), number(math.NaN()), 0, ErrOperand},
		{"overflow addition", "add", number(math.MaxFloat64), number(math.MaxFloat64), 0, ErrResult},
		{"overflow division", "divide", number(1), number(math.SmallestNonzeroFloat64), 0, ErrResult},
		{"overflow power", "power", number(10), number(309), 0, ErrResult},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Calculate(tt.operation, tt.a, tt.b)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("result = %v, want %v", got, tt.want)
			}
		})
	}
}
