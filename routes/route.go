// Package routes is for the routes
package routes

import (
	"net/http"

	"github.com/Arindam-langer/governance-service/handlers"
	"github.com/Arindam-langer/governance-service/internal/db"
	"github.com/Arindam-langer/governance-service/middleware"
)

// Init initializes the http server mux and registers routes
func Init(h *handlers.Handler, blockStore db.BlockStore) http.Handler {
	mux := http.NewServeMux()

	// Middleware instantiation
	blocklist := middleware.BlocklistMiddleware(blockStore)

	mux.HandleFunc("GET /health", h.HealthCheck)
	mux.HandleFunc("POST /signin", h.SignIn)
	mux.HandleFunc("POST /signup", h.SignUp)
	mux.HandleFunc("POST /refresh", h.Refresh)

	// Protected routes
	mux.Handle("POST /verify/token", blocklist(http.HandlerFunc(h.VerifyToken)))
	mux.Handle("POST /signout", blocklist(http.HandlerFunc(h.SignOut)))

	return mux
}
