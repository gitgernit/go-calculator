package agent

import (
	"context"
	"github.com/gitgernit/go-calculator/internal/domain/calculator"
	"github.com/gitgernit/go-calculator/internal/domain/orchestrator"
	"github.com/google/uuid"
	"log/slog"
	"sync"
)

var CalculatorInteractor = calculator.NewCalculatorInteractor()

type Interactor struct {
	poller ExpressionPoller
	mutex  sync.RWMutex
}

type ExpressionPoller interface {
	GetNextExpression() *orchestrator.Expression
	SolveExpression(id uuid.UUID, result float64) error
}

func (i *Interactor) StartPolling(context context.Context, workers int) error {
	for _ = range workers {
		go func() {
			err := i.SolveExpressions(context)
			if err != nil {
				slog.Error("error while solving expression: %v", err)
			}
		}()
	}

	return nil
}

func (i *Interactor) SolveExpressions(context context.Context) error {
	for {
		select {
		case <-context.Done():
			return context.Err()

		default:
			expression := i.poller.GetNextExpression()

			result, err := CalculatorInteractor.CalculateTokenized(expression.Tokens)
			if err != nil {
				return err
			}

			err = i.poller.SolveExpression(expression.Id, result)
			if err != nil {
				return err
			}
		}
	}
}
