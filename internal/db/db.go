// Package db is for db connection
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type UserStore interface{}
type DBStore struct {
	conn *pgx.Conn
}

func New(connectionURL string) (*DBStore, error) {
	conn, err := pgx.Connect(context.Background(), connectionURL)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}
	return &DBStore{conn: conn}, nil
}

func (d *DBStore) Close() {
	_ = d.conn.Close(context.Background())
}
