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
}

func NewGRPCPoller(host, port string) (*GRPCPoller, error) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%s", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &GRPCPoller{
		client: proto.NewOrchestratorServiceClient(conn),
		conn:   conn,
	}, nil
}

func (p *GRPCPoller) Close() error {
	return p.conn.Close()
}

func (p *GRPCPoller) GetNextTask(ctx context.Context) *agent.Task {
	stream, err := p.client.GetTasks(ctx)
	if err != nil {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			task, err := stream.Recv()
			if err != nil {
				return nil
			}

			id, err := uuid.Parse(task.Id)
			if err != nil {
				continue
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
	stream, err := p.client.GetTasks(context.Background())
	if err != nil {
		return err
	}

	resultFloat, _ := strconv.ParseFloat(result.Value, 64)
	return stream.Send(&proto.TaskResult{
		Id:     id.String(),
		Result: float32(resultFloat),
	})
}
