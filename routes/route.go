// Package routes is for the routes
package routes

import (
	"net/http"

	"github.com/Arindam-langer/authn-service/handlers"
	"github.com/Arindam-langer/authn-service/middleware"
)

// Init initializes the http server mux and registers routes
func Init(h *handlers.Handler, checker middleware.TokenBlockChecker) http.Handler {
	mux := http.NewServeMux()

	// Middleware instantiation
	blocklist := middleware.BlocklistMiddleware(checker)

	mux.HandleFunc("GET /health", h.HealthCheck)
	mux.HandleFunc("POST /signin", h.SignIn)
	mux.HandleFunc("POST /signup", h.SignUp)
	mux.HandleFunc("POST /refresh", h.Refresh)

	// Protected routes
	mux.Handle("POST /verify/token", blocklist(http.HandlerFunc(h.VerifyToken)))
	mux.Handle("POST /signout", blocklist(http.HandlerFunc(h.SignOut)))

	return mux
}
