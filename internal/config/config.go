package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all the configuration for the application.
type Config struct {
	ListenAddr   string
	DatabaseURL  string
	RedisURL     string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Load loads the environment variables and validates them.
func Load() (*Config, error) {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	cfg := &Config{}

	cfg.ListenAddr = os.Getenv("LISTEN_ADDR")
	if cfg.ListenAddr == "" {
		return nil, fmt.Errorf("LISTEN_ADDR environment variable is required")
	}

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	cfg.RedisURL = os.Getenv("REDIS_URL")
	if cfg.RedisURL == "" {
		return nil, fmt.Errorf("REDIS_URL environment variable is required")
	}

	readTimeoutStr := os.Getenv("READ_TIMEOUT")
	if readTimeoutStr == "" {
		return nil, fmt.Errorf("READ_TIMEOUT environment variable is required")
	}
	var err error
	cfg.ReadTimeout, err = time.ParseDuration(readTimeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid READ_TIMEOUT duration: %w", err)
	}

	writeTimeoutStr := os.Getenv("WRITE_TIMEOUT")
	if writeTimeoutStr == "" {
		return nil, fmt.Errorf("WRITE_TIMEOUT environment variable is required")
	}
	cfg.WriteTimeout, err = time.ParseDuration(writeTimeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid WRITE_TIMEOUT duration: %w", err)
	}

	return cfg, nil
}
