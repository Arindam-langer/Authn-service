// Package handlers:  functions for governance come here
package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Arindam-langer/governance-service/internal/auth"
	"github.com/Arindam-langer/governance-service/internal/db"
)

type Handler struct {
	store     db.UserStore
	authStore db.AuthStore
}

func New(store db.UserStore, authStore db.AuthStore) *Handler {
	return &Handler{store: store, authStore: authStore}
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

	accessToken, err := auth.CreateAccessToken(user.ID)
	if err != nil {
		slog.Error("error in generating access token", "error", err)
		throwError(w, "failed to generate token", http.StatusInternalServerError, err)
		return
	}
	refreshToken, err := auth.CreateRefreshToken()
	if err != nil {
		slog.Error("error in generating refresh Token", "error", err)
		throwError(w, "failed to generate Refresh Token", http.StatusInternalServerError, err)
		return
	}

	refreshExpiry := time.Now().Add(7 * 24 * time.Hour)
	sum := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(sum[:])

	http.SetCookie(w, &http.Cookie{
		Name:     "RefreshToken",
		HttpOnly: true,
		Value:    refreshToken,
		Secure:   true,
		Path:     "/refresh",
		SameSite: http.SameSiteStrictMode, // prevents csrf this one
		Expires:  refreshExpiry,
	})
	err = h.authStore.AddRefreshToken(req.Context(), user.ID, tokenHash, refreshExpiry)
	if err != nil {
		slog.Error("failed to add the add refresh token to user", "error", err)
		throwError(w, "failed to handle refresh token", http.StatusInternalServerError, err)
	}
	res := loginResponse{Code: 201}
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
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

func (h *Handler) RefreshToken(w http.ResponseWriter, req *http.Request) {
	// to refresh access token we get the RefreshToken from the cookies and check if RefreshToken is valid or not then we go from that
	// if token is not valid then we just return error and make the user sign in again i think this is better it means signin
	// has been done if the token is valid we just call create token and go from here

	// first we get the RefreshToken from the cookies of our  Request
	cookies, err := req.Cookie("RefreshToken")
	if err != nil {
		throwError(w, "failed to get cookies", http.StatusBadRequest, err)
		return
	}
	sum := sha256.Sum256([]byte(cookies.Value))
	tokenHash := hex.EncodeToString(sum[:])

	// cookies.Value == to the one in db then we are gucci
	token, err := h.authStore.GetRefreshToken(req.Context(), tokenHash)
	if err != nil {
		throwError(w, "Failed to fetch token", http.StatusInternalServerError, err)
		return
	}
	if token == nil {
		throwError(w, "Invalid refresh token", http.StatusUnauthorized, nil)
		return
	}
	if token.ExpiresAt.Before(time.Now()) {
		throwError(w, "Token Expired", http.StatusUnauthorized, nil)
		return
	}
	if token.Revoked {
		throwError(w, "Unauthorized token", http.StatusForbidden, nil)
		return
	}
	accessToken, err := auth.CreateAccessToken(token.UserID)
	if err != nil {
		slog.Error("error in generating access token", "error", err)
		throwError(w, "failed to generate token", http.StatusInternalServerError, err)
		return
	}
	res := loginResponse{Code: 201}
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	encode(w, res, http.StatusCreated)
}
