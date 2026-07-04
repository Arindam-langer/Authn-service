package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient is the concrete implementation of BlockStore using Redis.
type RedisClient struct {
	client *redis.Client
}
type BlockStore interface {
	BlockToken(ctx context.Context, tokenHash string, ttl time.Duration) error
	IsTokenBlocked(ctx context.Context, tokenHash string) (bool, error)
}

// NewRedis initializes a connection to Redis and pings it to verify connection.
func NewRedis(url string) (*RedisClient, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &RedisClient{client: client}, nil
}

// Close closes the Redis client connection.
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// BlockToken blocks a token for the specified TTL.
func (r *RedisClient) BlockToken(ctx context.Context, tokenHash string, ttl time.Duration) error {
	key := fmt.Sprintf("blocklist:%s", tokenHash)
	return r.client.Set(ctx, key, "true", ttl).Err()
}

// IsTokenBlocked checks if a token has been blocked.
func (r *RedisClient) IsTokenBlocked(ctx context.Context, tokenHash string) (bool, error) {
	key := fmt.Sprintf("blocklist:%s", tokenHash)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}
