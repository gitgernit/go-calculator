package orchestrator

import (
	"fmt"
	"github.com/gitgernit/go-calculator/internal/domain/calculator"
	"github.com/google/uuid"
	"sync"
)

type Interactor struct {
	Expressions     map[uuid.UUID]Expression
	ExpressionQueue []Expression
	mutex           sync.RWMutex
}

func (i *Interactor) AddExpression(tokens []calculator.Token) {
	expression := NewExpression(tokens)

	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.Expressions[expression.Id] = expression
	i.ExpressionQueue = append(i.ExpressionQueue, expression)
}

func (i *Interactor) ListExpressions() []*Expression {
	expressions := make([]*Expression, 0, len(i.Expressions))

	i.mutex.RLock()
	defer i.mutex.RUnlock()

	for _, v := range i.Expressions {
		expressions = append(expressions, &v)
	}

	return expressions
}

func (i *Interactor) GetExpression(id uuid.UUID) *Expression {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	expression, ok := i.Expressions[id]

	if !ok {
		return nil
	}

	return &expression
}

func (i *Interactor) GetNextExpression() *Expression {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	queueLength := len(i.ExpressionQueue)

	if queueLength == 0 {
		return nil
	}

	var expression Expression
	expression, i.ExpressionQueue = i.ExpressionQueue[queueLength-1], i.ExpressionQueue[:queueLength-1]

	expression.Status = Processing
	i.Expressions[expression.Id] = expression

	return &expression
}

func (i *Interactor) SolveExpression(id uuid.UUID, result float64) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	expression, ok := i.Expressions[id]

	if !ok {
		return fmt.Errorf("no such expression found")
	}

	expression.Status = Done
	expression.Result = result

	i.Expressions[expression.Id] = expression

	return nil
}
