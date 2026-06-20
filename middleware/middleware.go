// Package middleware has middleware functions used in server
package middleware

import (
	"log"
	"net/http"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("started", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Println("completed", r.Method, r.URL.Path)
	})
}

func UpdateHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Testing", "new middleware")
		next.ServeHTTP(w, r)
	})
}
