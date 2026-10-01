package middleware

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimitConfig struct {
	Endpoint string
	Limit    int64
	Window   time.Duration
}

type RedisClientInterface interface {
	IsConnected(ctx context.Context) bool
	Close() error
	Increment(ctx context.Context, key string, ttl time.Duration) (int64, error)
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, keys ...string) error
}

type RateLimiter struct {
	redis                 RedisClientInterface
	configs               map[string]RateLimitConfig
	failOpen              bool
	circuitBreakerState   string
	lastCheck             time.Time
	circuitBreakerTimeout time.Duration
}

// NewRateLimiter creates a new rate limiter with predefined limits per endpoint
func NewRateLimiter(redis RedisClientInterface) *RateLimiter {
	return &RateLimiter{
		redis: redis,
		configs: map[string]RateLimitConfig{
			"login":     {Endpoint: "login", Limit: 5, Window: 60 * time.Second},
			"register":  {Endpoint: "register", Limit: 3, Window: 60 * time.Second},
			"check_in":  {Endpoint: "check_in", Limit: 60, Window: 24 * time.Hour},
			"check_out": {Endpoint: "check_out", Limit: 60, Window: 24 * time.Hour},
			"default":   {Endpoint: "default", Limit: 100, Window: 60 * time.Second},
		},
		failOpen:              true, // Fail open: allow request if Redis unavailable
		circuitBreakerState:   "connected",
		lastCheck:             time.Now(),
		circuitBreakerTimeout: 30 * time.Second,
	}
}

// RateLimitMiddleware returns a middleware function that enforces rate limiting for a specific endpoint
func (rl *RateLimiter) RateLimitMiddleware(endpoint string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client IP
			clientIP := rl.extractClientIP(r)

			// Get rate limit config
			config, ok := rl.configs[endpoint]
			if !ok {
				config = rl.configs["default"]
			}

			// Check rate limit
			key := fmt.Sprintf("rate_limit:%s:%s", endpoint, clientIP)
			count, err := rl.checkRateLimit(r.Context(), key, config)

			if err != nil {
				// Redis error
				if rl.failOpen {
					log.Printf("⚠️  Rate limit check failed (fail-open): %v, allowing request from %s", err, clientIP)
					// Set headers anyway for consistency
					w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.Limit))
					w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", config.Limit))
					w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(config.Window).Unix()))
					next.ServeHTTP(w, r)
					return
				} else {
					// Fail closed
					log.Printf("❌ Rate limit check failed (fail-closed), blocking request from %s: %v", clientIP, err)
					w.Header().Set("Content-Type", "application/json")
					http.Error(w, `{"error": "service unavailable"}`, http.StatusServiceUnavailable)
					return
				}
			}

			// Set rate limit headers
			remaining := config.Limit - count + 1
			if remaining < 0 {
				remaining = 0
			}
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.Limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(config.Window).Unix()))

			// Check if limit exceeded
			if count > config.Limit {
				log.Printf("⚠️  Rate limit exceeded: %s from %s (endpoint: %s, limit: %d, attempt: %d)",
					r.RequestURI, clientIP, endpoint, config.Limit, count)
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error": "too many requests"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// checkRateLimit increments the counter in Redis and returns the current count
func (rl *RateLimiter) checkRateLimit(ctx context.Context, key string, config RateLimitConfig) (int64, error) {
	count, err := rl.redis.Increment(ctx, key, config.Window)
	if err != nil {
		if err == redis.Nil {
			// Key doesn't exist, first request
			return 1, nil
		}
		return 0, err
	}
	return count, nil
}

// extractClientIP extracts the real client IP from the request,
// handling X-Forwarded-For, X-Real-IP headers and RemoteAddr with port stripping
func (rl *RateLimiter) extractClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxy scenarios)
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// Take the first IP (original client)
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}

	// Fall back to RemoteAddr (remove port)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// No port present or invalid format
		return r.RemoteAddr
	}

	return host
}

// IsRedisAvailable checks Redis connectivity and manages circuit breaker state
func (rl *RateLimiter) IsRedisAvailable(ctx context.Context) bool {
	available := rl.redis.IsConnected(ctx)

	// Update circuit breaker state
	if available {
		rl.circuitBreakerState = "connected"
	} else {
		if time.Since(rl.lastCheck) > rl.circuitBreakerTimeout {
			rl.circuitBreakerState = "disconnected"
			log.Println("⚠️  Redis circuit breaker: OPEN (disconnected for >30s)")
		}
	}

	rl.lastCheck = time.Now()
	return available
}

// GetCircuitBreakerState returns the current circuit breaker state
func (rl *RateLimiter) GetCircuitBreakerState() string {
	return rl.circuitBreakerState
}
