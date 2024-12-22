package http

import "net/http"

func NewHTTPServer() (*http.Server, error) {
	router := http.NewServeMux()

	router.HandleFunc("/api/v1/calculate", CalculateHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	return server, nil
}
