package main

import (
	"context"
	appconfig "github.com/gitgernit/go-calculator/internal/config"
	"github.com/gitgernit/go-calculator/internal/domain/agent"
	grpcagent "github.com/gitgernit/go-calculator/internal/transport/grpc/agent"
	"strconv"
)

func main() {
	config, err := appconfig.New()
	if err != nil {
		panic(err)
	}

	poller, err := grpcagent.NewGRPCPoller(config.OrchestratorHost, strconv.Itoa(config.OrchestratorGRPCPort))
	if err != nil {
		panic(err)
	}
	defer poller.Close()

	interactor := agent.Interactor{
		Poller: poller,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = interactor.StartPolling(ctx, config.ComputingPower)
	if err != nil {
		panic(err)
	}

	select {}
}
