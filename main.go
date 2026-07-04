package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Arindam-langer/governance-service/handlers"
	"github.com/Arindam-langer/governance-service/internal/config"
	"github.com/Arindam-langer/governance-service/internal/db"
	"github.com/Arindam-langer/governance-service/middleware"
	"github.com/Arindam-langer/governance-service/routes"
)

func main() {
	// Initialize default global slog JSON logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Connect to postgres database
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()
	store, err := db.New(dbCtx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		slog.Info("closing database connection")
		store.Close()
	}()

	// Connect to redis
	redisStore, err := db.NewRedis(cfg.RedisURL)
	if err != nil {
		slog.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer func() {
		slog.Info("closing redis connection")
		redisStore.Close()
	}()

	// Initialize handlers, routes, and middleware chain
	h := handlers.New(store, store, redisStore)
	router := routes.Init(h, redisStore)
	
	globalRateLimit := middleware.RateLimitMiddleware(redisStore, 100, time.Minute)
	chain := middleware.RecoveryMiddleware(middleware.LoggingMiddleware(middleware.UpdateHeader(globalRateLimit(router))))

	// Setup Server
	srv := &http.Server{
		Addr:           cfg.ListenAddr,
		Handler:        chain,
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Start HTTP server
	go func() {
		slog.Info("server starting", "addr", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interruption signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("signal received, shutting down...")

	// Perform graceful shutdown with a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server exited cleanly")
}
