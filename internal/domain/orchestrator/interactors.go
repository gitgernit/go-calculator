package orchestrator

import (
	"fmt"
	"github.com/gitgernit/go-calculator/internal/domain/calculator"
	"github.com/google/uuid"
	"slices"
	"strconv"
	"sync"
)

var CalculatorInteractor = calculator.NewCalculatorInteractor()

type Task struct {
	Expression Expression
	Blocked    bool
	RPN        []calculator.Token
}

type Interactor struct {
	Expressions map[uuid.UUID]Expression
	TaskQueue   []*Task
	mutex       sync.RWMutex
}

func NewOrchestratorInteractor() *Interactor {
	return &Interactor{
		Expressions: make(map[uuid.UUID]Expression, 0),
		TaskQueue:   make([]*Task, 0),
	}
}

func (i *Interactor) AddExpression(tokens []calculator.Token) uuid.UUID {
	expression := NewExpression(tokens)

	i.mutex.Lock()
	defer i.mutex.Unlock()

	task := Task{
		Expression: expression,
		Blocked:    false,
		RPN:        CalculatorInteractor.TokenizedInfixToPolish(tokens),
	}

	i.Expressions[expression.Id] = expression
	i.TaskQueue = append(i.TaskQueue, &task)

	return expression.Id
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

	arg1Index, _, operationIndex, _, _, _, found := task.NextStep()
	if !found {
		return fmt.Errorf("no operation found in RPN")
	}

	token := calculator.Token{
		Value: fmt.Sprintf("%v", result),
	}

	task.RPN = append(task.RPN[:arg1Index], task.RPN[operationIndex+1:]...)
	task.RPN = append(task.RPN[:arg1Index], append([]calculator.Token{token}, task.RPN[arg1Index:]...)...)
	task.Blocked = false

	if len(task.RPN) == 1 {
		finalResult, err := strconv.ParseFloat(task.RPN[0].Value, 64)
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

func (t *Task) NextStep() (arg1Index, arg2Index, operationIndex int, arg1, arg2, operation string, found bool) {
	stack := []int{}
	operators := []string{"+", "-", "*", "/"}

	for i, token := range t.RPN {
		if slices.Contains(operators, token.Value) {
			if len(stack) < 2 {
				return -1, -1, -1, "", "", "", false
			}
			arg2Index = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			arg1Index = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			operationIndex = i
			arg1 = t.RPN[arg1Index].Value
			arg2 = t.RPN[arg2Index].Value
			operation = token.Value
			return arg1Index, arg2Index, operationIndex, arg1, arg2, operation, true
		}

		if _, err := strconv.Atoi(token.Value); err == nil {
			stack = append(stack, i)
		}
	}

	return -1, -1, -1, "", "", "", false
}
