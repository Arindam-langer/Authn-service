// Package db is for db connection
package db

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
)

type (
	User struct {
		ID        int
		Username  string
		Email     string
		PhoneUUID string
		Password  string
	}
	RefreshToken struct {
		ID        uuid.UUID
		UserID    int
		TokenHash string
		ExpiresAt time.Time
		CreatedAt time.Time
		Revoked   bool
	}
	UserStore interface {
		CreateUser(ctx context.Context, username, email, phoneUUID, password string) error
		GetUserByPhoneUUID(ctx context.Context, phoneUUID string) (*User, error)
	}
	AuthStore interface {
		AddRefreshToken(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error
		GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error)
		RevokeRefreshToken(ctx context.Context, tokenHash string) error
		RevokeAllUserTokens(ctx context.Context, userID int) error
	}
	DBStore struct {
		conn *pgx.Conn
	}
)

//go:embed schema.sql
var schema string

func New(connectionURL string) (*DBStore, error) {
	conn, err := pgx.Connect(context.Background(), connectionURL)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}
	_, err = conn.Exec(context.Background(), schema)
	if err != nil {
		return nil, fmt.Errorf("executing schema: %w", err)
	}
	return &DBStore{conn: conn}, nil
}

func (d *DBStore) Close() {
	_ = d.conn.Close(context.Background())
}

func (d *DBStore) CreateUser(ctx context.Context, username, email, phoneUUID, password string) error {
	_, err := d.conn.Exec(ctx, `
        INSERT INTO users (username, email, phone_uuid, password)
        VALUES ($1, $2, $3, $4)
    `, username, email, phoneUUID, password)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return nil
}

func (d *DBStore) GetUserByPhoneUUID(ctx context.Context, phoneUUID string) (*User, error) {
	var u User
	err := d.conn.QueryRow(ctx, `
		SELECT id, username, email, phone_uuid, password FROM users WHERE phone_uuid = $1
	`, phoneUUID).Scan(&u.ID, &u.Username, &u.Email, &u.PhoneUUID, &u.Password)
	if err != nil {
		return nil, fmt.Errorf("get user by phone uuid: %w", err)
	}
	return &u, nil
}

func (d *DBStore) AddRefreshToken(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
	_, err := d.conn.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("add refresh token: %w", err)
	}
	return nil
}

func (d *DBStore) GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	var rt RefreshToken
	err := d.conn.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, created_at, revoked
		 FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt, &rt.Revoked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get the refresh_tokens: %w", err)
	}

	return &rt, nil
}

func (d *DBStore) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := d.conn.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = TRUE WHERE token_hash = $1`,
		tokenHash)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func (d *DBStore) RevokeAllUserTokens(ctx context.Context, userID int) error {
	_, err := d.conn.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1`,
		userID)
	if err != nil {
		return fmt.Errorf("revoke all user tokens: %w", err)
	}
	return nil
}
