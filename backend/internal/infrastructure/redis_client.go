package infrastructure

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client          *redis.Client
	ctx             context.Context
	addr            string
	isConnected     bool
	mu              sync.RWMutex
	reconnectTicker *time.Ticker
	stopReconnect   chan bool
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
	rc := &RedisClient{
		client:        rdb,
		ctx:           context.Background(),
		addr:          addr,
		isConnected:   true,
		stopReconnect: make(chan bool),
	}
	// Start background reconnection routine
	go rc.startBackgroundReconnect()
	return rc, nil
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

	rc := &RedisClient{
		client:        rdb,
		ctx:           context.Background(),
		addr:          addr,
		isConnected:   true,
		stopReconnect: make(chan bool),
	}
	// Start background reconnection routine for tests too
	go rc.startBackgroundReconnect()
	return rc
}

// IsConnected checks if Redis is currently available
func (rc *RedisClient) IsConnected(ctx context.Context) bool {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.isConnected
}

// startBackgroundReconnect starts a goroutine that attempts to reconnect to Redis every 10 seconds
func (rc *RedisClient) startBackgroundReconnect() {
	rc.reconnectTicker = time.NewTicker(10 * time.Second)
	defer rc.reconnectTicker.Stop()

	for {
		select {
		case <-rc.stopReconnect:
			log.Println("Stopping Redis background reconnect routine")
			return
		case <-rc.reconnectTicker.C:
			rc.checkAndReconnect()
		}
	}
}

// checkAndReconnect checks and updates the connection state
func (rc *RedisClient) checkAndReconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rc.mu.Lock()
	defer rc.mu.Unlock()

	wasConnected := rc.isConnected

	// Try to ping Redis
	err := rc.client.Ping(ctx).Err()

	if err != nil && wasConnected {
		// Connection lost
		log.Printf("⚠ Redis connection lost: %v", err)
		rc.isConnected = false
	} else if err == nil && !wasConnected {
		// Connection restored
		log.Println("✓ Redis connection restored")
		rc.isConnected = true
	}
}

// Close gracefully closes the Redis connection
func (rc *RedisClient) Close() error {
	log.Println("Closing Redis connection...")
	// Stop the background reconnect routine
	close(rc.stopReconnect)
	return rc.client.Close()
}

// Increment increments a counter in Redis with TTL
func (rc *RedisClient) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	count, err := rc.client.Incr(ctx, key).Result()
	if err != nil {
		rc.updateConnectionState(false)
		return 0, err
	}

	// On successful operation, mark as connected
	rc.updateConnectionState(true)

	// Set TTL on first increment
	if count == 1 {
		rc.client.Expire(ctx, key, ttl)
	}

	return count, nil
}

// Get retrieves a value from Redis
func (rc *RedisClient) Get(ctx context.Context, key string) (string, error) {
	result, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		rc.updateConnectionState(false)
		return "", err
	}

	// On successful operation, mark as connected
	rc.updateConnectionState(true)
	return result, nil
}

// Delete removes one or more keys from Redis
func (rc *RedisClient) Delete(ctx context.Context, keys ...string) error {
	err := rc.client.Del(ctx, keys...).Err()
	if err != nil {
		rc.updateConnectionState(false)
		return err
	}

	// On successful operation, mark as connected
	rc.updateConnectionState(true)
	return nil
}

// updateConnectionState safely updates the isConnected flag
func (rc *RedisClient) updateConnectionState(connected bool) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if connected && !rc.isConnected {
		log.Println("✓ Redis connection restored (via operation)")
		rc.isConnected = true
	} else if !connected && rc.isConnected {
		log.Println("⚠ Redis connection lost (via operation)")
		rc.isConnected = false
	}
}

// GetClient returns the underlying Redis client
func (rc *RedisClient) GetClient() *redis.Client {
	return rc.client
}
