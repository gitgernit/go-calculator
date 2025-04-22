package orchestrator

import (
	"github.com/gitgernit/go-calculator/internal/domain/calculator"
	"github.com/google/uuid"
)

type Status int

const (
	Accepted Status = iota
	Done
)

type Expression struct {
	Id     uuid.UUID
	Owner  string
	Status Status
	Tokens []calculator.Token
	Result float64
}

func NewExpression(owner string, tokens []calculator.Token) Expression {
	return Expression{
		Id:     uuid.New(),
		Owner:  owner,
		Status: Accepted,
		Tokens: tokens,
		Result: 0.0,
	}
}
