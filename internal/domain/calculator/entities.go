package calculator

import (
	"strconv"
)

type Token struct {
	value string
}

func (t *Token) BinaryOperation(other Token) (operand1, operand2 float64, err error) {
	val1, err1 := strconv.ParseFloat(t.value, 64)
	val2, err2 := strconv.ParseFloat(other.value, 64)

	if err1 != nil || err2 != nil {
		if err1 != nil {
			return 0.0, 0.0, err1
		}

		return 0.0, 0.0, err2
	}

	return val1, val2, nil
}

func (t *Token) Sum(other Token) (Token, error) {
	val1, val2, err := t.BinaryOperation(other)

	if err != nil {
		return Token{"0.0"}, err
	}

	result := val1 + val2
	return Token{strconv.FormatFloat(result, 'f', -1, 64)}, nil
}

func (t *Token) Sub(other Token) (Token, error) {
	val1, val2, err := t.BinaryOperation(other)

	if err != nil {
		return Token{"0.0"}, err
	}

	result := val1 - val2
	return Token{strconv.FormatFloat(result, 'f', -1, 64)}, nil
}

func (t *Token) Mul(other Token) (Token, error) {
	val1, val2, err := t.BinaryOperation(other)

	if err != nil {
		return Token{"0.0"}, err
	}

	result := val1 * val2
	return Token{strconv.FormatFloat(result, 'f', -1, 64)}, nil
}

func (t *Token) Div(other Token) (Token, error) {
	val1, val2, err := t.BinaryOperation(other)

	if err != nil {
		return Token{"0.0"}, err
	}

	result := val1 / val2
	return Token{strconv.FormatFloat(result, 'f', -1, 64)}, nil
}
