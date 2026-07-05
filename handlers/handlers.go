// Package handlers:  functions for authn come here
package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Arindam-langer/authn-service/internal/auth"
)

type Handler struct {
	store      UserStore
	authStore  AuthStore
	blockStore BlockStore
}

func New(store UserStore, authStore AuthStore, blockStore BlockStore) *Handler {
	return &Handler{store: store, authStore: authStore, blockStore: blockStore}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, req *http.Request) {
	res := health{Message: "all good", Code: 200}
	encode(w, res, http.StatusOK)
}

func (h *Handler) SignIn(w http.ResponseWriter, req *http.Request) {
	var body loginRequest
	err := decode(req, &body)
	if err != nil {
		throwError(w, "invalid body payload", http.StatusBadRequest, err)
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
	err = h.issueRefreshToken(w, req.Context(), user.ID)
	if err != nil {
		slog.Error("failed to issue refresh token", "error", err)
		throwError(w, "failed to handle refresh token", http.StatusInternalServerError, err)
		return
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
		throwError(w, "invalid body payload", http.StatusBadRequest, err)
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

func (h *Handler) Refresh(w http.ResponseWriter, req *http.Request) {
	cookies, err := req.Cookie(refreshTokenCookieName)
	if err != nil {
		throwError(w, "failed to get cookies", http.StatusBadRequest, err)
		return
	}
	tokenHash := hashToken(cookies.Value)

	token, err := h.authStore.GetRefreshToken(req.Context(), tokenHash)
	if err != nil {
		throwError(w, "Failed to fetch token", http.StatusInternalServerError, err)
		return
	}
	if token == nil {
		throwError(w, "Invalid refresh token", http.StatusUnauthorized, nil)
		return
	}

	// Reuse Detection: If the token is already revoked, revoke ALL user sessions
	if token.Revoked {
		slog.Warn("Revoked refresh token reuse detected! Revoking all sessions for user", "user_id", token.UserID)
		_ = h.authStore.RevokeAllUserTokens(req.Context(), token.UserID)

		clearRefreshTokenCookie(w)
		throwError(w, "Nice Try!! Bitcch did you think this would word. Please log in again.", http.StatusUnauthorized, nil)
		return
	}

	// Expiry Check
	if token.ExpiresAt.Before(time.Now()) {
		throwError(w, "Token Expired", http.StatusUnauthorized, nil)
		return
	}

	// Generate new tokens
	accessToken, err := auth.CreateAccessToken(token.UserID)
	if err != nil {
		slog.Error("error in generating access token", "error", err)
		throwError(w, "failed to generate token", http.StatusInternalServerError, err)
		return
	}

	err = h.issueRefreshToken(w, req.Context(), token.UserID)
	if err != nil {
		slog.Error("failed to issue new refresh token", "error", err)
		throwError(w, "failed to handle refresh token", http.StatusInternalServerError, err)
		return
	}

	err = h.authStore.RevokeRefreshToken(req.Context(), tokenHash)
	if err != nil {
		slog.Error("failed to revoke old refresh token", "error", err)
	}

	res := loginResponse{Code: 201}
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	encode(w, res, http.StatusCreated)
}

func (h *Handler) issueRefreshToken(w http.ResponseWriter, ctx context.Context, userID int) error {
	refreshToken, err := auth.CreateRefreshToken()
	if err != nil {
		return fmt.Errorf("create token: %w", err)
	}

	refreshExpiry := time.Now().Add(7 * 24 * time.Hour)
	tokenHash := hashToken(refreshToken)

	err = h.authStore.AddRefreshToken(ctx, userID, tokenHash, refreshExpiry)
	if err != nil {
		return fmt.Errorf("db store: %w", err)
	}

	setRefreshTokenCookie(w, refreshToken, refreshExpiry)
	return nil
}

func (h *Handler) SignOut(w http.ResponseWriter, req *http.Request) {
	authHeader := req.Header["Authorization"]
	if len(authHeader) == 0 {
		throwError(w, "No token", http.StatusUnauthorized, nil)
		return
	}
	headerToken := strings.TrimPrefix(authHeader[0], "Bearer ")

	headerTokenHash := hashToken(headerToken)
	err := h.blockStore.BlockToken(req.Context(), headerTokenHash, 10*time.Minute)
	if err != nil {
		slog.Error("failed to block access token in blocklist", "error", err)
	}

	cookies, err := req.Cookie(refreshTokenCookieName)
	if err != nil {
		throwError(w, "failed to get cookies", http.StatusBadRequest, err)
		return
	}
	tokenHash := hashToken(cookies.Value)

	err = h.authStore.RevokeRefreshToken(req.Context(), tokenHash)
	if err != nil {
		throwError(w, "failed to revoke the token", http.StatusInternalServerError, err)
		return
	}

	clearRefreshTokenCookie(w)

	encode(w, struct {
		Message string `json:"message"`
	}{Message: "Signed Out successfully"}, http.StatusOK)
}
