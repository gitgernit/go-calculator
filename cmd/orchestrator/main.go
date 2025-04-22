package main

import (
	"net/http"
	"strconv"
	"sync"

	appconfig "github.com/gitgernit/go-calculator/internal/config"
	"github.com/gitgernit/go-calculator/internal/domain/orchestrator"
	"github.com/gitgernit/go-calculator/internal/infra/gorm"
	httporchestrator "github.com/gitgernit/go-calculator/internal/transport/http/orchestrator"
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

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		httpServer := httporchestrator.NewHTTPServer(interactor, config.OrchestratorHost, strconv.Itoa(config.OrchestratorPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
	}()

	wg.Wait()
}
