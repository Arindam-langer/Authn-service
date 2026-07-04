package handlers

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Arindam-langer/governance-service/internal/model"
)

// UserStore defines the user-related behavior handlers needs.
type UserStore interface {
	CreateUser(ctx context.Context, username, email, phoneUUID, password string) error
	GetUserByPhoneUUID(ctx context.Context, phoneUUID string) (*model.User, error)
}

// AuthStore defines the refresh-token behavior handlers needs.
type AuthStore interface {
	AddRefreshToken(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserTokens(ctx context.Context, userID int) error
}

// BlockStore defines the token-blocking behavior handlers needs.
type BlockStore interface {
	BlockToken(ctx context.Context, tokenHash string, ttl time.Duration) error
}

type (
	statusCode int
	health     struct {
		Message string     `json:"message"`
		Code    statusCode `json:"code"`
	}
	loginResponse struct {
		Code statusCode `json:"status"`
	}
	signUpRequest struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phone_number"`
		Password    string `json:"password"`
	}
	loginRequest struct {
		PhoneNumber string `json:"phone_number"`
		Password    string `json:"password"`
	}
)

func (r *signUpRequest) Validate() error {
	if r.Username == "" {
		return errors.New("username is required")
	}
	if r.Email == "" {
		return errors.New("email is required")
	}
	atIdx := strings.Index(r.Email, "@")
	if atIdx == -1 || atIdx == 0 || atIdx == len(r.Email)-1 {
		return errors.New("invalid email format: must contain @ followed by domain")
	}
	if r.PhoneNumber == "" {
		return errors.New("phone number is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

func (r *loginRequest) Validate() error {
	if r.PhoneNumber == "" {
		return errors.New("phone number is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}
