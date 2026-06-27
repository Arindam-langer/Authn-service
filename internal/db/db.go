// Package db is for db connection
package db

import (
	"context"
	_ "embed"
	"fmt"

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
	UserStore interface {
		CreateUser(ctx context.Context, username, email, phoneUUID, password string) error
		GetUserByPhoneUUID(ctx context.Context, phoneUUID string) (*User, error)
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
