// Package routes is for the routes
package routes

import (
	"net/http"

	"github.com/Arindam-langer/governance-service/handlers"
)

func Init(h *handlers.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.HealthCheck)
	mux.HandleFunc("POST /signin", h.SignIn)
	mux.HandleFunc("POST /verify/token", h.VerifyToken)
	return mux
}
