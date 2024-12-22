package main

import "github.com/gitgernit/go-calculator/internal/transport/http"

func main() {
	server, err := http.NewHTTPServer()
	if err != nil {
		panic(err)
	}

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
