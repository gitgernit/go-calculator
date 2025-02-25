package main

import (
	"github.com/gitgernit/go-calculator/internal/transport/http/calc"
)

// Obsolete -- "calc" is the solution to previous sprint's final task.

func main() {
	server, err := calc.NewHTTPServer()
	if err != nil {
		panic(err)
	}

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
