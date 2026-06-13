package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Arindam-langer/governance-service/handlers"
)

const (
	PORT         string        = "8080"
	ReadTimeout  time.Duration = 10 * time.Second
	WriteTimeout time.Duration = 10 * time.Second
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.HealthCheck)
	s := &http.Server{
		Addr:           PORT,
		Handler:        mux,
		ReadTimeout:    ReadTimeout,
		WriteTimeout:   WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Println("server listening on localhost:8080")
	err := s.ListenAndServe()
	if err != nil {
		log.Fatalf("failed to run the server")
	}
}
