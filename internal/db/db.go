// Package db is for db connection
package db

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type (
	UserStore interface {
		CreateUser(ctx context.Context, userName, password string) error
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

func (d *DBStore) CreateUser(ctx context.Context, userName, password string) error {
	_, err := d.conn.Exec(ctx, `
        INSERT INTO users (username, password)
        VALUES ($1, $2)
    `, userName, password)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return nil
}
