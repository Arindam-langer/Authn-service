// Package db is for db connection
package db

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

type (
	UserStore interface{}
	DBStore   struct {
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
		log.Fatalf("error executing the script %v", err)
	}
	return &DBStore{conn: conn}, nil
}

func (d *DBStore) Close() {
	_ = d.conn.Close(context.Background())
}

// do we pass a whole user struct we will create or just userName
func (d *DBStore) CreateUser() {
}
