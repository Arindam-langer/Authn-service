// Package model defines shared domain types used across packages.
package model

import (
	"time"

	"github.com/gofrs/uuid"
)

type User struct {
	ID        int
	Username  string
	Email     string
	PhoneUUID string
	Password  string
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    int
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	Revoked   bool
}
