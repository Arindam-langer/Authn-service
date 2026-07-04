// Package db is for db connection
package db

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/Arindam-langer/authn-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBStore struct {
	pool *pgxpool.Pool
}

//go:embed schema.sql
var schema string

// New initializes a connection pool to Postgres and verifies the connection.
func New(ctx context.Context, connectionURL string) (*DBStore, error) {
	pool, err := pgxpool.New(ctx, connectionURL)
	if err != nil {
		return nil, fmt.Errorf("db connect pool: %w", err)
	}

	// Ping to verify connection is alive
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	_, err = pool.Exec(ctx, schema)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("executing schema: %w", err)
	}
	return &DBStore{pool: pool}, nil
}

// Close closes the connection pool.
func (d *DBStore) Close() {
	d.pool.Close()
}

func (d *DBStore) CreateUser(ctx context.Context, username, email, phoneUUID, password string) error {
	_, err := d.pool.Exec(ctx, `
        INSERT INTO users (username, email, phone_uuid, password)
        VALUES ($1, $2, $3, $4)
    `, username, email, phoneUUID, password)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return nil
}

func (d *DBStore) GetUserByPhoneUUID(ctx context.Context, phoneUUID string) (*model.User, error) {
	var u model.User
	err := d.pool.QueryRow(ctx, `
		SELECT id, username, email, phone_uuid, password FROM users WHERE phone_uuid = $1
	`, phoneUUID).Scan(&u.ID, &u.Username, &u.Email, &u.PhoneUUID, &u.Password)
	if err != nil {
		return nil, fmt.Errorf("get user by phone uuid: %w", err)
	}
	return &u, nil
}

func (d *DBStore) AddRefreshToken(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
	_, err := d.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("add refresh token: %w", err)
	}
	return nil
}

func (d *DBStore) GetRefreshToken(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	err := d.pool.QueryRow(ctx,
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
	_, err := d.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = TRUE WHERE token_hash = $1`,
		tokenHash)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func (d *DBStore) RevokeAllUserTokens(ctx context.Context, userID int) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1`,
		userID)
	if err != nil {
		return fmt.Errorf("revoke all user tokens: %w", err)
	}
	return nil
}
