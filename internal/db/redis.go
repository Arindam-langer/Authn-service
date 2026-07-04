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

// Allow implements a fixed-window rate limiting check using Redis.
func (r *RedisClient) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now().UnixNano()
	bucket := now / int64(window)
	redisKey := fmt.Sprintf("ratelimit:%s:%d", key, bucket)

	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, window*2)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("redis ratelimit pipeline: %w", err)
	}

	count := incr.Val()
	return count <= int64(limit), nil
}

