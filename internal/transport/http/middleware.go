package http

import (
	"encoding/json"
	"net/http"
)

type Middleware func(next http.Handler) http.Handler

func PanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				response := map[string]string{"error": "Internal server error"}
				err := json.NewEncoder(w).Encode(response)
				if err != nil {
					panic(err)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func CreateStackedMiddleware(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}

		return next
	}
}
