package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gitgernit/go-calculator/internal/config"
	"github.com/gitgernit/go-calculator/internal/domain/agent"
	"github.com/gitgernit/go-calculator/internal/domain/calculator"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Task struct {
	ID              uuid.UUID `json:"id"`
	Arg1            string    `json:"arg1"`
	Arg2            string    `json:"arg2"`
	Operation       string    `json:"operation"`
	OperationTimeMS int       `json:"operation_time"`
}

type TaskResponse struct {
	Task *Task `json:"task"`
}

type ExpressionPoller struct {
	Config config.Config
}

func (p *ExpressionPoller) GetNextTask(context context.Context) *agent.Task {
	url := fmt.Sprintf("http://%s:%d/internal/task", p.Config.OrchestratorHost, p.Config.OrchestratorPort)
	for {
		select {
		case <-context.Done():
			return nil

		default:
			req, err := http.NewRequestWithContext(context, "GET", url, nil)
			if err != nil {
				fmt.Println("error creating request:", err)
				return nil
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println("error making request:", err)
				return nil
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				time.Sleep(time.Duration(p.Config.PollingIntervalMS) * time.Millisecond)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				fmt.Println("unexpected response status:", resp.Status)
				return nil
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("error reading response:", err)
				return nil
			}

			var taskResponse TaskResponse
			if err := json.Unmarshal(body, &taskResponse); err != nil {
				fmt.Println("error decoding JSON:", err)
				return nil
			}

			return &agent.Task{
				ID:              taskResponse.Task.ID,
				Arg1:            calculator.Token{Value: taskResponse.Task.Arg1},
				Arg2:            calculator.Token{Value: taskResponse.Task.Arg2},
				Operation:       calculator.Token{Value: taskResponse.Task.Operation},
				OperationTimeMS: taskResponse.Task.OperationTimeMS,
			}
		}
	}
}

func (p *ExpressionPoller) SolveTask(id uuid.UUID, result calculator.Token) error {
	url := fmt.Sprintf("http://%s:%d/internal/task", p.Config.OrchestratorHost, p.Config.OrchestratorPort)
	resultFloat, _ := strconv.ParseFloat(result.Value, 64)
	payload := map[string]interface{}{
		"id":     id,
		"result": resultFloat,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send result, status: %v", resp.Status)
	}

	return nil
}
