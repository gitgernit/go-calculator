package calc

import (
	transporthttp "github.com/gitgernit/go-calculator/internal/transport/http"
	"net/http"
)

func NewHTTPServer() (*http.Server, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/calculate", CalculateHandler)

	stack := transporthttp.CreateStackedMiddleware(
		transporthttp.PanicMiddleware,
	)
	handler := stack(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	return server, nil
}
