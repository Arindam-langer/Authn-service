// Package middleware has middleware functions used in server
package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Arindam-langer/governance-service/internal/db"
)

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

func BlocklistMiddleware(blockStore db.BlockStore) func(http.Handler) http.Handler {
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

			blocked, err := blockStore.IsTokenBlocked(r.Context(), headerTokenHash)
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
