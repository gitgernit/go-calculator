package orchestrator

import (
	"encoding/json"
	"fmt"
	"github.com/gitgernit/go-calculator/internal/config"
	"github.com/gitgernit/go-calculator/internal/domain/auth"
	"github.com/gitgernit/go-calculator/internal/domain/calculator"
	"github.com/gitgernit/go-calculator/internal/domain/orchestrator"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"sync"
)

var CalculatorInteractor = calculator.NewCalculatorInteractor()
var Config, _ = config.New()
var AuthInteractor = auth.UserInteractor{JWTSecretKey: Config.JWTSecretKey}

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

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	owner, err := AuthInteractor.CheckToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
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

	id := s.Interactor.AddExpression(owner, tokens)

	resp := ExpressionResponse{ID: id}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) ListExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	owner, err := AuthInteractor.CheckToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	expressions, err := s.Interactor.ListExpressions(owner)
	if err != nil {
		http.Error(w, "Failed to fetch expressions", http.StatusInternalServerError)
		return
	}

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

func (s *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	err := AuthInteractor.Create(req.Login, req.Password)
	if err != nil {
		if err.Error() == "user already exists" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "user created"})
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	token, err := AuthInteractor.Authorize(req.Login, req.Password)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err.Error() == "invalid password" {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			http.Error(w, "Failed to authorize", http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func NewHTTPServer(interactor *orchestrator.Interactor, host string, port string) *http.Server {
	srv := &Server{Interactor: interactor}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/calculate", srv.AddExpressionHandler)
	mux.HandleFunc("/api/v1/expressions", srv.ListExpressionsHandler)
	mux.HandleFunc("/api/v1/expressions/", srv.GetExpressionHandler)
	mux.HandleFunc("/api/v1/register", srv.RegisterHandler)
	mux.HandleFunc("/api/v1/login", srv.LoginHandler)
	//mux.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
	//	switch r.Method {
	//	case http.MethodGet:
	//		srv.GetTaskHandler(w, r)
	//
	//	case http.MethodPost:
	//		srv.SolveTaskHandler(w, r)
	//
	//	default:
	//		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	//	}
	//})

	return &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, port),
		Handler: mux,
	}
}
