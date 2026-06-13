// Package routes is for the routes
package routes

import (
	"net/http"

	"github.com/Arindam-langer/governance-service/handlers"
)

func Init() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handlers.HealthCheck)
	return mux
}
