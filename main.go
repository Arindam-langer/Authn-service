package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Arindam-langer/governance-service/middleware"
	"github.com/Arindam-langer/governance-service/routes"
)

const (
	listenAddr   string        = "localhost:8080"
	ReadTimeout  time.Duration = 10 * time.Second
	WriteTimeout time.Duration = 10 * time.Second
)

func main() {
	router := routes.Init()

	handlers := middleware.LoggingMiddleware(middleware.UpdateHeader(router))
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
