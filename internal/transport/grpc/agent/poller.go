package agent

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gitgernit/go-calculator/internal/domain/agent"
	"github.com/gitgernit/go-calculator/internal/domain/calculator"
	"github.com/gitgernit/go-calculator/internal/transport/grpc/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCPoller struct {
	client proto.OrchestratorServiceClient
	conn   *grpc.ClientConn
	stream proto.OrchestratorService_GetTasksClient
}

func NewGRPCPoller(host, port string) (*GRPCPoller, error) {
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%s", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := proto.NewOrchestratorServiceClient(conn)

	stream, err := client.GetTasks(context.Background())
	if err != nil {
		return nil, err
	}

	return &GRPCPoller{
		client: client,
		conn:   conn,
		stream: stream,
	}, nil
}

func (p *GRPCPoller) Close() error {
	return p.conn.Close()
}

func (p *GRPCPoller) GetNextTask(ctx context.Context) *agent.Task {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			task, err := p.stream.Recv()
			if err != nil {
				return nil
			}

			id, err := uuid.Parse(task.Id)
			if err != nil {
				return nil
			}

			return &agent.Task{
				ID:              id,
				Arg1:            calculator.Token{Value: task.Arg1},
				Arg2:            calculator.Token{Value: task.Arg2},
				Operation:       calculator.Token{Value: task.Operation},
				OperationTimeMS: int(task.OperationTime),
			}
		}
	}
}

func (p *GRPCPoller) SolveTask(id uuid.UUID, result calculator.Token) error {
	resultFloat, err := strconv.ParseFloat(result.Value, 64)
	if err != nil {
		return err
	}

	return p.stream.Send(&proto.TaskResult{
		Id:     id.String(),
		Result: float32(resultFloat),
	})
}
