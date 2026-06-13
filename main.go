package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Arindam-langer/governance-service/routes"
)

const (
	listenAddr   string        = "localhost:8080"
	ReadTimeout  time.Duration = 10 * time.Second
	WriteTimeout time.Duration = 10 * time.Second
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("started", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Println("completed", r.Method, r.URL.Path)
	})
}

func main() {
	router := routes.Init()

	handlers := loggingMiddleware(router)
	s := &http.Server{
		Addr:           listenAddr,
		Handler:        handlers,
		ReadTimeout:    ReadTimeout,
		WriteTimeout:   WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Println("server listening on ", listenAddr)
	err := s.ListenAndServe()
	if err != nil {
		log.Fatalf("failed to run the server %v", err)
	}
}
