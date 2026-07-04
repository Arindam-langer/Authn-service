// Package middleware has middleware functions used in server
package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// This file was the consmer so we add the interface here.
type TokenBlockChecker interface {
	IsTokenBlocked(ctx context.Context, tokenHash string) (bool, error)
}

// RateLimiter defines the behavior needed to evaluate rate limit rules.
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("request started", "method", r.Method, "path", r.URL.Path)

		next.ServeHTTP(w, r)

		slog.Info("request completed", "method", r.Method, "path", r.URL.Path)
	})
}

func UpdateHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Testing", "new middleware")
		next.ServeHTTP(w, r)
	})
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "error", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"message":"internal server error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func BlocklistMiddleware(checker TokenBlockChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"message":"No token"}`))
				return
			}

			headerToken := strings.TrimPrefix(authHeader, "Bearer ")
			sumHeader := sha256.Sum256([]byte(headerToken))
			headerTokenHash := hex.EncodeToString(sumHeader[:])

			blocked, err := checker.IsTokenBlocked(r.Context(), headerTokenHash)
			if err != nil {
				slog.Error("failed to check blocklist", "error", err)
			} else if blocked {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"message":"token is revoked"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware enforces rate limiting per client IP per route.
func RateLimitMiddleware(limiter RateLimiter, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			// Rate limit per client IP, HTTP Method, and Path for route isolation.
			key := fmt.Sprintf("ip:%s:route:%s:%s", ip, r.Method, r.URL.Path)

			allowed, err := limiter.Allow(r.Context(), key, limit, window)
			if err != nil {
				// Fail-open: we log the error but allow the request so Redis issues don't block users.
				slog.Error("rate limiter evaluation error", "error", err, "ip", ip, "path", r.URL.Path)
			} else if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"message":"Too many requests. Please try again later."}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}
