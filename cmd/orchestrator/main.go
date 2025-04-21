package main

import (
	appconfig "github.com/gitgernit/go-calculator/internal/config"
	"github.com/gitgernit/go-calculator/internal/domain/orchestrator"
	"github.com/gitgernit/go-calculator/internal/infra/gorm"
	httporchestrator "github.com/gitgernit/go-calculator/internal/transport/http/orchestrator"
	"strconv"
)

func main() {
	config, err := appconfig.New()
	if err != nil {
		panic(err)
	}

	err = gorm.Initialize()
	if err != nil {
		panic(err)
	}

	interactor := orchestrator.NewOrchestratorInteractor()
	server := httporchestrator.NewHTTPServer(interactor, config.OrchestratorHost, strconv.Itoa(config.OrchestratorPort))

	err = server.ListenAndServe()

	if err != nil {
		panic(err)
	}
}
