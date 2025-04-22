generate-grpc:
	protoc --go_out=. --go-grpc_out=. --grpc-gateway_out . ./internal/transport/grpc/proto/orchestrator.proto