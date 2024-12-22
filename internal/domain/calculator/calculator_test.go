package calculator

import (
	"testing"
)

func TestCalculate(t *testing.T) {
	interactor := NewCalculatorInteractor()

	tests := []struct {
		expression string
		expected   float64
		err        bool
	}{
		{"3+5", 8.0, false},
		{"10-2*3", 4.0, false},
		{"(1+2)*(3+4)", 21.0, false},
		{"10/2+3", 8.0, false},
		{"(4+5)*(2-1)", 9.0, false},
		{"3+(2*(4-1))", 9.0, false},
		{"3+", 0.0, true},
		{"*3+5", 0.0, true},
		{"1+(1+(1+(1))", 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			result, err := interactor.Calculate(tt.expression)

			if (err != nil) != tt.err {
				t.Errorf("expected error: %v, got: %v", tt.err, err)
			}
			if !tt.err && result != tt.expected {
				t.Errorf("expected result: %v, got: %v", tt.expected, result)
			}
		})
	}
}
