package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func encode(w http.ResponseWriter, response any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func throwError(w http.ResponseWriter, message string, statusCode int, err error) {
	if err != nil {
		slog.Error(message, "error", err)
	} else {
		slog.Warn(message)
	}
	encode(w, struct {
		Message string `json:"message"`
	}{Message: message}, statusCode)
}

func decode(req *http.Request, body any) error {
	return json.NewDecoder(req.Body).Decode(body)
}
