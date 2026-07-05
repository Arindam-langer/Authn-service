package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

const (
	refreshTokenCookieName = "RefreshToken"
	refreshTokenPath       = "/refresh"
)

// hashToken hashes a token using SHA-256 to store or compare safely
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// setRefreshTokenCookie sets the secure http-only refresh cookie
func setRefreshTokenCookie(w http.ResponseWriter, value string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    value,
		Path:     refreshTokenPath,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  expiresAt,
	})
}

// clearRefreshTokenCookie clears the refresh cookie on the client side
func clearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     refreshTokenPath,
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1,
	})
}

func encode(w http.ResponseWriter, response any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func throwError(w http.ResponseWriter, message string, statusCode int, err error) {
	var maxBytesErr *http.MaxBytesError
	if err != nil && errors.As(err, &maxBytesErr) {
		statusCode = http.StatusRequestEntityTooLarge
		message = "request body too large"
	}

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
