package infrastructure

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisClient initializes a Redis client with connection pooling, timeout, and retry logic
func NewRedisClient(addr string) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		MaxRetries:   3,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	log.Println("✓ Redis connected:", addr)
	return &RedisClient{client: rdb, ctx: context.Background()}, nil
}

// NewRedisClientForTesting creates a Redis client for testing (used in tests)
func NewRedisClientForTesting(addr string) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		MaxRetries:   1,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	return &RedisClient{client: rdb, ctx: context.Background()}
}

// IsConnected checks if Redis is currently available
func (rc *RedisClient) IsConnected(ctx context.Context) bool {
	if err := rc.client.Ping(ctx).Err(); err != nil {
		return false
	}
	return true
}

// Close gracefully closes the Redis connection
func (rc *RedisClient) Close() error {
	log.Println("Closing Redis connection...")
	return rc.client.Close()
}

// Increment increments a counter in Redis with TTL
func (rc *RedisClient) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	count, err := rc.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// Set TTL on first increment
	if count == 1 {
		rc.client.Expire(ctx, key, ttl)
	}

	return count, nil
}

// Get retrieves a value from Redis
func (rc *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return rc.client.Get(ctx, key).Result()
}

// Delete removes one or more keys from Redis
func (rc *RedisClient) Delete(ctx context.Context, keys ...string) error {
	return rc.client.Del(ctx, keys...).Err()
}

// GetClient returns the underlying Redis client
func (rc *RedisClient) GetClient() *redis.Client {
	return rc.client
}
