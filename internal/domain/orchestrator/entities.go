package orchestrator

import (
	"github.com/gitgernit/go-calculator/internal/domain/calculator"
	"github.com/google/uuid"
)

type Status int

const (
	Accepted Status = iota
	Processing
	Done
)

type Expression struct {
	Id     uuid.UUID
	Status Status
	Tokens []calculator.Token
	Result float64
}

func NewExpression(tokens []calculator.Token) Expression {
	return Expression{
		Id:     uuid.New(),
		Status: Accepted,
		Tokens: tokens,
		Result: 0.0,
	}
}
