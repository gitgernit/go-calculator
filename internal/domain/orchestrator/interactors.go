package orchestrator

import (
	"fmt"
	"github.com/gitgernit/go-calculator/internal/domain/calculator"
	db "github.com/gitgernit/go-calculator/internal/infra/gorm"
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
	TaskQueue []*Task
	mutex     sync.RWMutex
}

func NewOrchestratorInteractor() *Interactor {
	interactor := &Interactor{
		TaskQueue: make([]*Task, 0),
	}

	if err := interactor.loadPendingExpressions(); err != nil {
		panic(fmt.Sprintf("failed to load pending expressions: %v", err))
	}

	return interactor
}

func (i *Interactor) loadPendingExpressions() error {
	var expressions []db.Expression
	err := db.Db.Where("status != ?", db.Done).Find(&expressions).Error
	if err != nil {
		return err
	}

	i.mutex.Lock()
	defer i.mutex.Unlock()

	for _, dbExpr := range expressions {
		tokens := toTokenSlice(dbExpr.Tokens)
		expr := Expression{
			Id:     dbExpr.ID,
			Owner:  dbExpr.Owner,
			Status: Status(dbExpr.Status),
			Tokens: tokens,
			Result: dbExpr.Result,
		}

		task := &Task{
			Expression: expr,
			Blocked:    false,
			RPN:        CalculatorInteractor.TokenizedInfixToPolish(tokens),
		}

		i.TaskQueue = append(i.TaskQueue, task)
	}

	return nil
}

func (i *Interactor) AddExpression(owner string, tokens []calculator.Token) uuid.UUID {
	expression := NewExpression(owner, tokens)
	db.Db.Create(&db.Expression{
		ID:     expression.Id,
		Owner:  expression.Owner,
		Status: db.Status(expression.Status),
		Tokens: toStringSlice(expression.Tokens),
		Result: expression.Result,
	})

	i.mutex.Lock()
	defer i.mutex.Unlock()

	task := Task{
		Expression: expression,
		Blocked:    false,
		RPN:        CalculatorInteractor.TokenizedInfixToPolish(tokens),
	}

	i.TaskQueue = append(i.TaskQueue, &task)

	return expression.Id
}

func (i *Interactor) ListExpressions(owner string) ([]*Expression, error) {
	var dbExpressions []db.Expression
	if err := db.Db.Where("owner = ?", owner).Find(&dbExpressions).Error; err != nil {
		return nil, err
	}

	expressions := make([]*Expression, len(dbExpressions))
	for idx, e := range dbExpressions {
		expressions[idx] = &Expression{
			Id:     e.ID,
			Owner:  e.Owner,
			Status: Status(e.Status),
			Tokens: toTokenSlice(e.Tokens),
			Result: e.Result,
		}
	}

	return expressions, nil
}

func (i *Interactor) GetExpression(id uuid.UUID) *Expression {
	var e db.Expression
	if err := db.Db.First(&e, "id = ?", id).Error; err != nil {
		return nil
	}

	return &Expression{
		Id:     e.ID,
		Owner:  e.Owner,
		Status: Status(e.Status),
		Tokens: toTokenSlice(e.Tokens),
		Result: e.Result,
	}
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

		var expr db.Expression
		if err := db.Db.First(&expr, "id = ?", id).Error; err != nil {
			return fmt.Errorf("failed to find expression: %v", err)
		}

		expr.Status = db.Done
		expr.Result = finalResult

		if err := db.Db.Save(&expr).Error; err != nil {
			return fmt.Errorf("failed to update expression: %v", err)
		}
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

func toStringSlice(tokens []calculator.Token) []string {
	strs := make([]string, len(tokens))
	for i, t := range tokens {
		strs[i] = t.Value
	}
	return strs
}

func toTokenSlice(values []string) []calculator.Token {
	strs := make([]calculator.Token, len(values))
	for i, t := range values {
		strs[i] = calculator.Token{Value: t}
	}
	return strs
}
