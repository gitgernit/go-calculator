package main

import (
	"context"
	appconfig "github.com/gitgernit/go-calculator/internal/config"
	"github.com/gitgernit/go-calculator/internal/domain/agent"
	httpagent "github.com/gitgernit/go-calculator/internal/transport/http/agent"
)

func main() {
	config, err := appconfig.New()

	if err != nil {
		panic(err)
	}

	poller := httpagent.ExpressionPoller{
		Config: *config,
	}
	interactor := agent.Interactor{
		Poller: &poller,
	}

	ctx, cancel := context.WithCancel(context.Background())
	err = interactor.StartPolling(ctx, config.ComputingPower)

	if err != nil {
		cancel()
		panic(err)
	}
}
