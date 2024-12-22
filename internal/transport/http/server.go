package http

import "net/http"

func NewHTTPServer() (*http.Server, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/calculate", CalculateHandler)

	stack := CreateStackedMiddleware(
		PanicMiddleware,
	)
	handler := stack(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	return server, nil
}
