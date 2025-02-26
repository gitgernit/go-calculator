package orchestrator

import (
	"encoding/json"
	"fmt"
	"github.com/gitgernit/go-calculator/internal/config"
	"github.com/gitgernit/go-calculator/internal/domain/calculator"
	"github.com/gitgernit/go-calculator/internal/domain/orchestrator"
	"github.com/google/uuid"
	"net/http"
	"sync"
)

var CalculatorInteractor = calculator.NewCalculatorInteractor()
var Config, _ = config.New()

type Server struct {
	Interactor *orchestrator.Interactor
	Mutex      sync.Mutex
}

type ExpressionRequest struct {
	Expression string `json:"expression"`
}

type ExpressionResponse struct {
	ID uuid.UUID `json:"id"`
}

type ExpressionVerboseResponse struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

type ExpressionsListResponse struct {
	Expressions []ExpressionVerboseResponse `json:"expressions"`
}

type TaskResponse struct {
	ID            uuid.UUID `json:"id"`
	Arg1          string    `json:"arg1"`
	Arg2          string    `json:"arg2"`
	Operation     string    `json:"operation"`
	OperationTime int       `json:"operation_time"`
}

type TaskResultRequest struct {
	ID     uuid.UUID `json:"id"`
	Result float64   `json:"result"`
}

func (s *Server) AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExpressionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusUnprocessableEntity)
		return
	}

	tokens, err := CalculatorInteractor.TokenizeInfix(req.Expression)
	if err != nil {
		http.Error(w, "Invalid expression", http.StatusUnprocessableEntity)
		return
	}

	id := s.Interactor.AddExpression(tokens)

	resp := ExpressionResponse{ID: id}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) ListExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	expressions := s.Interactor.ListExpressions()
	resp := ExpressionsListResponse{Expressions: make([]ExpressionVerboseResponse, 0)}
	for _, expr := range expressions {
		var status string

		if expr.Status == orchestrator.Accepted {
			status = "accepted"
		} else {
			status = "done"
		}

		resp.Expressions = append(resp.Expressions, ExpressionVerboseResponse{
			ID:     expr.Id.String(),
			Status: status,
			Result: expr.Result,
		})
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *Server) GetExpressionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := r.URL.Path[len("/api/v1/expressions/"):]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	expr := s.Interactor.GetExpression(id)
	if expr == nil {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	var status string

	if expr.Status == orchestrator.Accepted {
		status = "accepted"
	} else {
		status = "done"
	}

	json.NewEncoder(w).Encode(map[string]*ExpressionVerboseResponse{
		"expression": &ExpressionVerboseResponse{
			ID:     expr.Id.String(),
			Status: status,
			Result: expr.Result,
		},
	})
}

func (s *Server) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	task := s.Interactor.GetNextTask()
	if task == nil {
		http.Error(w, "No task available", http.StatusNotFound)
		return
	}

	_, _, _, arg1, arg2, operation, _ := task.NextStep()
	var executionTime int

	switch operation {
	case "+":
		executionTime = Config.TimeAdditionMS
	case "-":
		executionTime = Config.TimeSubtractionMS
	case "*":
		executionTime = Config.TimeMultiplicationsMS
	case "/":
		executionTime = Config.TimeDivisionsMS
	}

	resp := struct {
		Task TaskResponse `json:"task"`
	}{
		Task: TaskResponse{
			ID:            task.Expression.Id,
			Arg1:          arg1,
			Arg2:          arg2,
			Operation:     operation,
			OperationTime: executionTime,
		},
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *Server) SolveTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TaskResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusUnprocessableEntity)
		return
	}

	if err := s.Interactor.SolveTask(req.ID, req.Result); err != nil {
		if err.Error() == "no such task found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func NewHTTPServer(interactor *orchestrator.Interactor, host string, port string) *http.Server {
	srv := &Server{Interactor: interactor}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/calculate", srv.AddExpressionHandler)
	mux.HandleFunc("/api/v1/expressions", srv.ListExpressionsHandler)
	mux.HandleFunc("/api/v1/expressions/", srv.GetExpressionHandler)
	mux.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			srv.GetTaskHandler(w, r)

		case http.MethodPost:
			srv.SolveTaskHandler(w, r)

		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	return &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, port),
		Handler: mux,
	}
}
