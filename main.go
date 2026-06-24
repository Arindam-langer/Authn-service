package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Arindam-langer/governance-service/handlers"
	"github.com/Arindam-langer/governance-service/internal/db"
	"github.com/Arindam-langer/governance-service/middleware"
	"github.com/Arindam-langer/governance-service/routes"
	"github.com/joho/godotenv"
)

const (
	listenAddr   string        = "localhost:8080"
	ReadTimeout  time.Duration = 10 * time.Second
	WriteTimeout time.Duration = 10 * time.Second
)

// graceful shutdown
func main() {
	_ = godotenv.Load()

	store, err := db.New(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("could not connect to db: %v", err)
	}
	defer store.Close()
	h := handlers.New(store)
	router := routes.Init(h)

	chain := middleware.LoggingMiddleware(middleware.UpdateHeader(router))
	s := &http.Server{
		Addr:           listenAddr,
		Handler:        chain,
		ReadTimeout:    ReadTimeout,
		WriteTimeout:   WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		log.Println("server starting on :8080")
		if err := s.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()
	// Main goroutine stays free
	quit := make(chan os.Signal, 1)
	signal.Notify(
		quit,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	<-quit
	log.Println("signal received, shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server exited cleanly")
}
