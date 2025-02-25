package calculator

import (
	"fmt"
	"slices"
	"strconv"
)

type Interactor struct{}

func NewCalculatorInteractor() *Interactor {
	return &Interactor{}
}

func (i *Interactor) Calculate(expression string) (float64, error) {
	tokenized, err := i.tokenizeInfix(expression)

	if err != nil {
		return 0.0, err
	}

	polish := i.tokenizedInfixToPolish(tokenized)
	result, err := i.solveRPN(polish)

	if err != nil {
		return 0.0, err
	}

	return result, nil
}

func (i *Interactor) CalculateTokenized(expression []Token) (float64, error) {
	polish := i.tokenizedInfixToPolish(expression)
	result, err := i.solveRPN(polish)

	if err != nil {
		return 0.0, err
	}

	return result, nil
}

func (i *Interactor) tokenizeInfix(infix string) ([]Token, error) {
	var result []Token
	var currString string

	for _, char := range infix {
		switch rune(char) {
		case '+', '-', '*', '/', '(', ')':
			if len(currString) > 0 {
				result = append(result, Token{currString})
				currString = ""
			}

			result = append(result, Token{string(char)})

		default:
			currString += string(char)
		}
	}

	if len(currString) > 0 {
		result = append(result, Token{currString})
	}

	err := i.validateTokenizedInfix(result)

	return result, err
}

func (i *Interactor) validateTokenizedInfixParentheses(infix []Token) error {
	var stack []Token

	for _, token := range infix {
		if token.value == "(" {
			stack = append(stack, token)
		} else if token.value == ")" {
			if len(stack) == 0 {
				return fmt.Errorf("expected an opening parenthesis")
			}

			parenthesis := stack[len(stack)-1].value
			stack = stack[:len(stack)-1]

			if parenthesis != "(" {
				return fmt.Errorf("expected an opening parenthesis")
			}
		}
	}

	if len(stack) > 0 {
		return fmt.Errorf("insufficient amount of parentheses")
	}

	return nil
}

func (i *Interactor) validateTokenizedInfix(infix []Token) error {
	binary := []string{"+", "-", "*", "/"}
	special := []string{"(", ")"}

	err := i.validateTokenizedInfixParentheses(infix)

	if err != nil {
		return err
	}

	for index := range len(infix) {
		value := infix[index].value

		if slices.Contains(binary, value) {
			if index == 0 {
				return fmt.Errorf("expected a first operand for a binary operator (%s)", value)
			}

			if index == len(infix)-1 {
				return fmt.Errorf("expected a second operand for a binary operator (%s)", value)
			}

			previousValue := infix[index-1].value
			nextValue := infix[index+1].value

			if slices.Contains(binary, nextValue) {
				return fmt.Errorf("expected a number or parentheses after a binary operator, got %s", nextValue)
			}

			if slices.Contains(binary, previousValue) {
				return fmt.Errorf("expected a number or parentheses before a binary operator, got %s", previousValue)
			}
		} else if !slices.Contains(special, value) {
			if index != 0 {
				previousValue := infix[index-1].value

				if !slices.Contains(binary, previousValue) && !slices.Contains(special, previousValue) {
					return fmt.Errorf("expected a binary operator or a parenthesis before a number, got %s", previousValue)
				}
			}

			if index != len(infix)-1 {
				nextValue := infix[index+1].value

				if !slices.Contains(binary, nextValue) && !slices.Contains(special, nextValue) {
					return fmt.Errorf("expected a binary operator or a parenthesis after a number, got %s", nextValue)
				}
			}
		}
	}

	return nil
}

func (i *Interactor) tokenizedInfixToPolish(infix []Token) []Token {
	priorities := map[string]int{
		"(": 0, ")": 0,
		"+": 1, "-": 1,
		"*": 2, "/": 2,
	}
	output, stack := make([]Token, 0, len(infix)), make([]Token, 0, len(infix))

	for _, token := range infix {
		switch token.value {
		case "+", "-", "*", "/":
			for len(stack) > 0 &&
				priorities[stack[len(stack)-1].value] >= priorities[token.value] {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}

			stack = append(stack, token)

		case "(":
			stack = append(stack, token)

		case ")":
			for stack[len(stack)-1].value != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}

			stack = stack[:len(stack)-1]

		default:
			output = append(output, token)
		}
	}

	for len(stack) > 0 {
		token := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if priorities[token.value] >= 1 {
			output = append(output, token)
		}
	}

	return output
}

func (i *Interactor) solveRPN(rpn []Token) (float64, error) {
	if len(rpn) == 0 {
		return 0.0, fmt.Errorf("received a blank reverse polish notation")
	}

	var stack []Token

	for _, token := range rpn {
		var newToken Token
		var err error

		if token.value == "+" {
			newToken, err = stack[len(stack)-2].Sum(stack[len(stack)-1])
		} else if token.value == "-" {
			newToken, err = stack[len(stack)-2].Sub(stack[len(stack)-1])
		} else if token.value == "*" {
			newToken, err = stack[len(stack)-2].Mul(stack[len(stack)-1])
		} else if token.value == "/" {
			newToken, err = stack[len(stack)-2].Div(stack[len(stack)-1])
		} else {
			newToken = token
		}

		if err != nil {
			return 0.0, err
		}

		if newToken != token {
			stack = stack[:len(stack)-2]
		}

		stack = append(stack, newToken)
	}

	result, err := strconv.ParseFloat(stack[0].value, 64)
	return result, err
}
