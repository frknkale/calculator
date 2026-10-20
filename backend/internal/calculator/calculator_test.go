package calculator

import (
	"errors"
	"strings"
	"testing"
)

func operand(value string) *string { return &value }

func TestCalculate(t *testing.T) {
	tests := []struct {
		name, operation string
		a, b            *string
		want            string
		wantErr         error
	}{
		{"addition", "add", operand("2"), operand("3"), "5", nil},
		{"exact decimal addition", "add", operand("0.1"), operand("0.2"), "0.3", nil},
		{"16 significant digits", "add", operand("1.2345678901234564"), operand("0"), "1.234567890123456", nil},
		{"positive significant tie", "add", operand("1.2345678901234565"), operand("0"), "1.234567890123457", nil},
		{"negative significant tie", "add", operand("-1.2345678901234565"), operand("0"), "-1.234567890123457", nil},
		{"rounding carries into exponent", "add", operand("9999999999999999.5"), operand("0"), "1e+16", nil},
		{"plain upper boundary", "add", operand("9999999999999999"), operand("0"), "9999999999999999", nil},
		{"scientific upper boundary", "add", operand("1e16"), operand("0"), "1e+16", nil},
		{"plain lower boundary", "add", operand("1e-6"), operand("0"), "0.000001", nil},
		{"scientific lower boundary", "add", operand("1e-7"), operand("0"), "1e-7", nil},
		{"negative scientific result", "add", operand("-1e25"), operand("0"), "-1e+25", nil},
		{"large integer with fraction", "add", operand("1e25"), operand("0.2"), "1e+25", nil},
		{"beyond safe integer", "add", operand("9007199254740992"), operand("1"), "9007199254740993", nil},
		{"negative decimal", "add", operand("-1.5"), operand("0.25"), "-1.25", nil},
		{"decimal syntax", "add", operand("+.5"), operand("2."), "2.5", nil},
		{"subtraction order", "subtract", operand("2"), operand("5"), "-3", nil},
		{"exact subtraction", "subtract", operand("0.3"), operand("0.1"), "0.2", nil},
		{"multiplication", "multiply", operand("-4"), operand("2.5"), "-10", nil},
		{"exact multiplication", "multiply", operand("0.1"), operand("0.2"), "0.02", nil},
		{"zero is present", "multiply", operand("7"), operand("0"), "0", nil},
		{"division", "divide", operand("7"), operand("2"), "3.5", nil},
		{"division with leading zeros", "divide", operand("1"), operand("300"), "0.003333333333333333", nil},
		{"division across exponent boundary", "divide", operand("9"), operand("10"), "0.9", nil},
		{"tiny recurring division", "divide", operand("1e-100"), operand("3"), "3.333333333333333e-101", nil},
		{"large recurring division", "divide", operand("1e100"), operand("3"), "3.333333333333333e+99", nil},
		{"division near rounding boundary", "divide", operand("1.2345678901234564999999999"), operand("1"), "1.234567890123456", nil},
		{"recurring division", "divide", operand("1"), operand("3"), "0." + strings.Repeat("3", 16), nil},
		{"division rounds up", "divide", operand("2"), operand("3"), "0." + strings.Repeat("6", 15) + "7", nil},
		{"positive tiny value", "divide", operand("5e-35"), operand("1"), "5e-35", nil},
		{"negative tiny value", "divide", operand("-5e-35"), operand("1"), "-5e-35", nil},
		{"tiny division stays nonzero", "divide", operand("1e-35"), operand("1"), "1e-35", nil},
		{"zero numerator", "divide", operand("0"), operand("2"), "0", nil},
		{"division by zero", "divide", operand("3"), operand("0"), "", ErrDivisionByZero},
		{"division by negative zero", "divide", operand("3"), operand("-0"), "", ErrDivisionByZero},
		{"power", "power", operand("2"), operand("10"), "1024", nil},
		{"large power", "power", operand("10"), operand("25"), "1e+25", nil},
		{"power beyond previous range", "power", operand("10"), operand("309"), "1e+309", nil},
		{"power limit boundary", "power", operand("1"), operand("1000"), "1", nil},
		{"negative power limit boundary", "power", operand("1"), operand("-1000"), "1", nil},
		{"negative exponent", "power", operand("2"), operand("-2"), "0.25", nil},
		{"reported small power", "power", operand("0.1"), operand("35"), "1e-35", nil},
		{"fractional power with leading zeros", "power", operand("0.000123456789"), operand("0.5"), "0.01111111106055556", nil},
		{"fractional power exponent correction", "power", operand("3.7489e-33"), operand("0.5"), "6.122826144845205e-17", nil},
		{"tiny fractional power", "power", operand("2e-100"), operand("0.5"), "1.414213562373095e-50", nil},
		{"negative exponent with leading zeros", "power", operand("2"), operand("-35"), "2.91038304567337e-11", nil},
		{"negative exponent near decimal cutoff", "power", operand("2.5"), operand("-40"), "1.208925819614629e-16", nil},
		{"tiny reciprocal power", "power", operand("3"), operand("-100"), "1.940325217482633e-48", nil},
		{"zero positive power", "power", operand("0"), operand("2"), "0", nil},
		{"fractional power", "power", operand("2"), operand("0.5"), "1.414213562373095", nil},
		{"negative base integer power", "power", operand("-2"), operand("3"), "-8", nil},
		{"tiny power stays nonzero", "power", operand("1e-35"), operand("1"), "1e-35", nil},
		{"zero to zero", "power", operand("0"), operand("0"), "1", nil},
		{"imaginary power", "power", operand("-1"), operand("0.5"), "", ErrResult},
		{"zero negative power", "power", operand("0"), operand("-1"), "", ErrResult},
		{"large exponent", "power", operand("2"), operand("1001"), "", ErrPowerLimit},
		{"large negative exponent", "power", operand("2"), operand("-1001"), "", ErrPowerLimit},
		{"huge exponent", "power", operand("2"), operand("1e100"), "", ErrPowerLimit},
		{"large scientific result", "power", operand("1e5"), operand("1000"), "1e+5000", nil},
		{"square root", "sqrt", operand("81"), nil, "9", nil},
		{"irrational root", "sqrt", operand("2"), nil, "1.414213562373095", nil},
		{"small root odd exponent", "sqrt", operand("0.00002"), nil, "0.004472135954999579", nil},
		{"small root even exponent", "sqrt", operand("0.0002"), nil, "0.01414213562373095", nil},
		{"large root", "sqrt", operand("1e50"), nil, "1e+25", nil},
		{"tiny root stays nonzero", "sqrt", operand("1e-70"), nil, "1e-35", nil},
		{"zero root", "sqrt", operand("0"), nil, "0", nil},
		{"negative root", "sqrt", operand("-1"), nil, "", ErrNegativeRoot},
		{"percentage", "percentage", operand("200"), operand("25"), "50", nil},
		{"decimal percentage", "percentage", operand("1234.56"), operand("12.5"), "154.32", nil},
		{"negative percentage", "percentage", operand("200"), operand("-50"), "-100", nil},
		{"negative percentage base", "percentage", operand("-200"), operand("25"), "-50", nil},
		{"zero percent", "percentage", operand("200"), operand("0"), "0", nil},
		{"zero percentage base", "percentage", operand("0"), operand("25"), "0", nil},
		{"over one hundred percent", "percentage", operand("200"), operand("150"), "300", nil},
		{"large percentage base", "percentage", operand("1e25"), operand("12.5"), "1.25e+24", nil},
		{"tiny exact percentage", "percentage", operand("1e-50"), operand("12.5"), "1.25e-51", nil},
		{"small exact multiplication", "multiply", operand("1e-30"), operand("1e-30"), "1e-60", nil},
		{"upper input exponent", "add", operand("1e1000"), operand("0"), "1e+1000", nil},
		{"lower input exponent", "add", operand("1e-1000"), operand("0"), "1e-1000", nil},
		{"unknown operation", "modulo", operand("1"), operand("2"), "", ErrOperation},
		{"missing operation", "", operand("1"), operand("2"), "", ErrOperation},
		{"missing a", "add", nil, operand("2"), "", ErrOperand},
		{"missing b", "add", operand("1"), nil, "", ErrOperand},
		{"missing percentage rate", "percentage", operand("200"), nil, "", ErrOperand},
		{"unexpected b", "sqrt", operand("4"), operand("2"), "", ErrOperand},
		{"empty a", "add", operand(""), operand("2"), "", ErrOperand},
		{"invalid b", "add", operand("1"), operand("NaN"), "", ErrOperand},
		{"infinite a", "add", operand("Infinity"), operand("2"), "", ErrOperand},
		{"hexadecimal", "sqrt", operand("0x10"), nil, "", ErrOperand},
		{"whitespace", "sqrt", operand(" 4 "), nil, "", ErrOperand},
		{"trailing newline", "sqrt", operand("4\n"), nil, "", ErrOperand},
		{"long operand", "sqrt", operand(strings.Repeat("9", 129)), nil, "", ErrOperand},
		{"input length boundary", "add", operand(strings.Repeat("9", 128)), operand("0"), "1e+128", nil},
		{"input exponent too high", "sqrt", operand("1e1001"), nil, "", ErrOperand},
		{"input exponent too low", "sqrt", operand("1e-1001"), nil, "", ErrOperand},
		{"unrepresentable exponent", "sqrt", operand("1e99999999999999999999"), nil, "", ErrOperand},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Calculate(tt.operation, tt.a, tt.b)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("result = %q, want %q", got, tt.want)
			}
		})
	}
}
