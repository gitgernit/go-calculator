package agent

import (
	"context"
	"fmt"
	"github.com/gitgernit/go-calculator/internal/domain/calculator"
	"github.com/google/uuid"
	"log/slog"
	"sync"
	"time"
)

var CalculatorInteractor = calculator.NewCalculatorInteractor()

type Interactor struct {
	Poller ExpressionPoller
	mutex  sync.RWMutex
}

type Task struct {
	ID              uuid.UUID        `json:"id"`
	Arg1            calculator.Token `json:"arg1"`
	Arg2            calculator.Token `json:"arg2"`
	Operation       calculator.Token `json:"operation"`
	OperationTimeMS int              `json:"operation_time"`
}

type ExpressionPoller interface {
	GetNextTask(context context.Context) *Task
	SolveTask(id uuid.UUID, result calculator.Token) error
}

func (i *Interactor) StartPolling(context context.Context, workers int) error {
	for _ = range workers {
		go func() {
			err := i.SolveTasks(context)
			if err != nil {
				slog.Error("error while solving expression: %v", err)
			}
		}()
	}

	return nil
}

func (i *Interactor) SolveTasks(context context.Context) error {
	for {
		select {
		case <-context.Done():
			return context.Err()

		default:
			task := i.Poller.GetNextTask(context)

			if task == nil {
				return fmt.Errorf("invalid task received")
			}

			time.Sleep(time.Duration(task.OperationTimeMS) * time.Millisecond)

			result, err := CalculatorInteractor.CalculateTokenized([]calculator.Token{task.Arg1, task.Arg2, task.Operation})
			if err != nil {
				return err
			}

			err = i.Poller.SolveTask(task.ID, calculator.Token{Value: fmt.Sprintf("%v", result)})
			if err != nil {
				return err
			}
		}
	}
}
