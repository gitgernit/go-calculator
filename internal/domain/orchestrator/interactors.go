package orchestrator

import (
	"fmt"
	"github.com/gitgernit/go-calculator/internal/domain/calculator"
	"github.com/google/uuid"
	"strconv"
	"sync"
)

type Task struct {
	Expression Expression
	Blocked    bool
}

type Interactor struct {
	Expressions map[uuid.UUID]Expression
	TaskQueue   []*Task
	mutex       sync.RWMutex
}

func (i *Interactor) AddExpression(tokens []calculator.Token) {
	expression := NewExpression(tokens)

	i.mutex.Lock()
	defer i.mutex.Unlock()

	task := Task{
		Expression: expression,
		Blocked:    false,
	}

	i.Expressions[expression.Id] = expression
	i.TaskQueue = append(i.TaskQueue, &task)
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

func (i *Interactor) GetNextTask() *Task {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	var task *Task

	for _, t := range i.TaskQueue {
		if !t.Blocked {
			task = t
			break
		}
	}

	if task == nil {
		return nil
	}

	task.Blocked = true

	return task
}

func (i *Interactor) SolveTask(id uuid.UUID, result float64) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	var taskIndex int
	var task *Task
	found := false

	for index, t := range i.TaskQueue {
		if t.Expression.Id == id {
			task = t
			taskIndex = index
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("no such task found")
	}

	token := calculator.Token{
		Value: fmt.Sprintf("%v", result),
	}

	task.Expression.Tokens = task.Expression.Tokens[3:]
	task.Expression.Tokens = append([]calculator.Token{token}, task.Expression.Tokens...)
	task.Blocked = false

	if len(task.Expression.Tokens) == 1 {
		finalResult, err := strconv.ParseFloat(task.Expression.Tokens[0].Value, 64)
		if err != nil {
			return fmt.Errorf("failed to parse final result: %v", err)
		}

		i.TaskQueue = append(i.TaskQueue[:taskIndex], i.TaskQueue[taskIndex+1:]...)

		expr := i.Expressions[id]
		expr.Status = Done
		expr.Result = finalResult
		i.Expressions[id] = expr
	}

	return nil
}
