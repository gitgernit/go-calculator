package orchestrator

import (
	"github.com/gitgernit/go-calculator/internal/config"
	"github.com/gitgernit/go-calculator/internal/domain/orchestrator"
	"sync"
	"time"

	"github.com/gitgernit/go-calculator/internal/transport/grpc/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

var Config, _ = config.New()

type Server struct {
	proto.UnimplementedOrchestratorServiceServer
	Interactor *orchestrator.Interactor
	Mutex      sync.Mutex
}

func NewServer(interactor *orchestrator.Interactor) *Server {
	return &Server{
		Interactor: interactor,
	}
}

func (s *Server) GetTasks(stream proto.OrchestratorService_GetTasksServer) error {
	go func() {
		for {
			result, err := stream.Recv()
			if err != nil {
				return
			}

			id, err := uuid.Parse(result.Id)
			if err != nil {
				continue
			}

			s.Mutex.Lock()
			err = s.Interactor.SolveTask(id, float64(result.Result))
			s.Mutex.Unlock()

			if err != nil {
				continue
			}
		}
	}()

	for {
		s.Mutex.Lock()
		task := s.Interactor.GetNextTask()
		s.Mutex.Unlock()

		if task == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		_, _, _, arg1, arg2, operation, _ := task.NextStep()
		var execTime uint64

		switch operation {
		case "+":
			execTime = uint64(Config.TimeAdditionMS)
		case "-":
			execTime = uint64(Config.TimeSubtractionMS)
		case "*":
			execTime = uint64(Config.TimeMultiplicationsMS)
		case "/":
			execTime = uint64(Config.TimeDivisionsMS)
		default:
			continue
		}

		err := stream.Send(&proto.IncomingTask{
			Id:            task.Expression.Id.String(),
			Arg1:          arg1,
			Arg2:          arg2,
			Operation:     operation,
			OperationTime: execTime,
		})

		if err != nil {
			return err
		}
	}
}

func RegisterService(grpcServer *grpc.Server, interactor *orchestrator.Interactor) {
	s := NewServer(interactor)
	proto.RegisterOrchestratorServiceServer(grpcServer, s)
}
