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

func decode(req *http.Request, body any) error {
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
	
	phoneUUID := auth.GeneratePhoneUUID(body.PhoneNumber)
	user, err := h.store.GetUserByPhoneUUID(req.Context(), phoneUUID)
	if err != nil {
		throwError(w, "user not found", http.StatusNotFound, err)
		return
	}
	
	if user.Password != body.Password {
		throwError(w, "invalid password", http.StatusUnauthorized, nil)
		return
	}

	token, err := auth.CreateToken(user.ID)
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
	if len(authHeader) == 0 {
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
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		throwError(w, "unverified", http.StatusForbidden, nil)
		return
	}
	
	userID := int(userIDFloat)
	log.Printf("Token verified for user ID: %d", userID)
	
	res := struct {
		Message string `json:"message"`
	}{"token verified"}
	encode(w, res, http.StatusOK)
}

func (h *Handler) SignUp(w http.ResponseWriter, req *http.Request) {
	var body signUpRequest
	err := decode(req, &body)
	if err != nil {
		throwError(w, "invalid body", http.StatusInternalServerError, err)
		return
	}

	// Validate email: must contain '@' and have characters before and after it
	atIdx := strings.Index(body.Email, "@")
	if atIdx == -1 || atIdx == 0 || atIdx == len(body.Email)-1 {
		throwError(w, "invalid email format: must contain @ followed by domain", http.StatusBadRequest, nil)
		return
	}

	phoneUUID := auth.GeneratePhoneUUID(body.PhoneNumber)
	err = h.store.CreateUser(req.Context(), body.Username, body.Email, phoneUUID, body.Password)
	if err != nil {
		throwError(w, "failed to create User", http.StatusInternalServerError, err)
		return
	}

	encode(w, struct {
		Message string `json:"message"`
	}{"user created"}, http.StatusCreated)
}
