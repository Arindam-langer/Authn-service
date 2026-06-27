package main

import (
	"context"
	"log/slog"
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

	// Initialize the default global slog logger to output JSON
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	store, err := db.New(os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("could not connect to db", "error", err)
		os.Exit(1)
	}
	defer store.Close()
	h := handlers.New(store)
	router := routes.Init(h)

	chain := middleware.RecoveryMiddleware(middleware.LoggingMiddleware(middleware.UpdateHeader(router)))
	s := &http.Server{
		Addr:           listenAddr,
		Handler:        chain,
		ReadTimeout:    ReadTimeout,
		WriteTimeout:   WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		slog.Info("server starting", "addr", listenAddr)
		if err := s.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
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
	slog.Info("signal received, shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
		os.Exit(1)
	}
	slog.Info("server exited cleanly")
}
