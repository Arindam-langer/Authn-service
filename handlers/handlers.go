// Package handlers:  functions for governance come here
package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Arindam-langer/governance-service/internal/auth"
	"github.com/Arindam-langer/governance-service/internal/db"
)

type Handler struct {
	store db.UserStore
}

func New(store db.UserStore) *Handler {
	return &Handler{store: store}
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

	if err := body.Validate(); err != nil {
		throwError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	phoneUUID := auth.GeneratePhoneUUID(body.PhoneNumber)
	user, err := h.store.GetUserByPhoneUUID(req.Context(), phoneUUID)
	if err != nil {
		throwError(w, "user not found", http.StatusNotFound, err)
		return
	}

	if !auth.CheckPasswordHash(body.Password, user.Password) {
		throwError(w, "invalid password", http.StatusUnauthorized, nil)
		return
	}

	token, err := auth.CreateToken(user.ID)
	if err != nil {
		slog.Error("error in generating token", "error", err)
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

	userID, err := auth.VerifyToken(headerToken)
	if err != nil {
		throwError(w, "invalid token", http.StatusUnauthorized, err)
		return
	}

	slog.Info("token verified successfully", "user_id", userID)

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

	if err := body.Validate(); err != nil {
		throwError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	hashedPassword, err := auth.HashPassword(body.Password)
	if err != nil {
		throwError(w, "failed to process password", http.StatusInternalServerError, err)
		return
	}

	phoneUUID := auth.GeneratePhoneUUID(body.PhoneNumber)
	err = h.store.CreateUser(req.Context(), body.Username, body.Email, phoneUUID, hashedPassword)
	if err != nil {
		throwError(w, "failed to create User", http.StatusInternalServerError, err)
		return
	}

	encode(w, struct {
		Message string `json:"message"`
	}{"user created"}, http.StatusCreated)
}
