package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func encode(w http.ResponseWriter, response any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error in handling response %v", err)
	}
}

func throwError(w http.ResponseWriter, message string, statusCode int, err error) {
	if err != nil {
		log.Printf("Error: %s: %v", message, err)
	} else {
		log.Printf("Error: %s", message)
	}
	encode(w, struct {
		Message string `json:"message"`
	}{Message: message}, statusCode)
}

func decode(req *http.Request, body any) error {
	return json.NewDecoder(req.Body).Decode(body)
}
