// Package handlers:  functions for governance come here
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Arindam-langer/governance-service/internal/auth"
	"github.com/Arindam-langer/governance-service/internal/db"
	"github.com/golang-jwt/jwt"
)

type Handler struct {
	store db.UserStore
}

func New(store db.UserStore) *Handler {
	return &Handler{store: store}
}

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

func decode(req *http.Request, body *loginRequest) error {
	return json.NewDecoder(req.Body).Decode(body)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, req *http.Request) {
	res := health{Message: "all good", Code: 200}
	encode(w, res, http.StatusOK)
}

func (h *Handler) SignIn(w http.ResponseWriter, req *http.Request) {
	var body loginRequest
	err := decode(req, &body)
	if err != nil {
		throwError(w, "invalid body", http.StatusInternalServerError, err)
		return
	}
	// this would be a DB call where we call db to check if this user exists or not for now it is going to be hard coded
	if body.Username != "aru" {
		throwError(w, "user not found", http.StatusNotFound, nil)
		return
	}
	// if the above condition does not work then we are gucci so we just generate the token and return in header
	token, err := auth.CreateToken(body.Username, body.Password)
	if err != nil {
		log.Printf("error in generating Token %v", err)
		throwError(w, "failed to generate token", http.StatusInternalServerError, err)
		return
	}

	res := loginResponse{Code: 201}
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	encode(w, res, http.StatusCreated)
}

func (h *Handler) VerifyToken(w http.ResponseWriter, req *http.Request) {
	authHeader := req.Header["Authorization"]
	if authHeader == nil || len(authHeader) == 0 {
		throwError(w, "No token", http.StatusUnauthorized, nil)
		return
	}
	headerToken := strings.TrimPrefix(authHeader[0], "Bearer ")
	token, err := jwt.Parse(headerToken, auth.IsValidToken)
	if err != nil {
		log.Printf("error in token parsing %v", err)
		throwError(w, "invalid token", http.StatusUnauthorized, err)
		return
	}
	if token == nil || !token.Valid {
		log.Printf("Not a valid Token")
		throwError(w, "invalid token", http.StatusUnauthorized, nil)
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		throwError(w, "invalid claims", http.StatusUnauthorized, nil)
		return
	}
	username, ok := claims["username"].(string)
	if !ok || username != "aru" {
		throwError(w, "unverified", http.StatusForbidden, nil)
		return
	}
	res := struct {
		Message string `json:"message"`
	}{"token verified"}
	encode(w, res, http.StatusOK)
}

func (h *Handler) SignUp(w http.ResponseWriter, req *http.Request) {
	var body loginRequest
	err := decode(req, &body)
	if err != nil {
		throwError(w, "invalid body", http.StatusInternalServerError, err)
		return
	}
	err = h.store.CreateUser(req.Context(), body.Username, body.Password)
	if err != nil {
		throwError(w, "failed to create User", http.StatusInternalServerError, err)
		return
	}

	encode(w, struct {
		Message string `json:"message"`
	}{"user created"}, http.StatusCreated)
}
