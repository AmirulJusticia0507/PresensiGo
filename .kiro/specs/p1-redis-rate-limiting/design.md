# P1 #3: Activate Redis Rate Limiting — Design

## Overview
Activate Redis rate limiting to protect against brute force attacks and DDoS. Initialize Redis connection with proper lifecycle management, wire rate-limit middleware to protected endpoints, normalize client IP identification, define per-endpoint rate limits, handle Redis unavailability gracefully, and add health check endpoint.

## Architecture
### Rate Limiting Flow
```
Request arrives
  ↓
[Rate Limit Middleware]
  ├─ Extract client IP
  ├─ Determine endpoint and rate limit
  ├─ Check Redis: key = "rate_limit:{endpoint}:{ip}"
  ├─ If count < limit:
  │   ├─ Increment count
  │   ├─ Set TTL (e.g., 60s for per-minute limits, 86400s for per-day)
  │   └─ → Allow request, continue handler
  ├─ If count >= limit:
  │   ├─ Return HTTP 429
  │   ├─ Add X-RateLimit headers
  │   └─ → Block request, log violation
  └─ If Redis error:
      ├─ If fail-open policy: log warning, allow request
      ├─ If fail-closed policy: return HTTP 503
      └─ Track circuit breaker state
```

## Components and Interfaces

### Redis Client Component
- **File:** `backend/internal/infrastructure/redis_client.go`
- **Purpose:** Manages Redis connection lifecycle, connection pooling, and basic Redis operations
- **Interfaces:** 
  - `NewRedisClient(addr string)`: Factory function to create Redis client
  - `Increment(ctx context.Context, key string, ttl time.Duration)`: Atomic increment with TTL
  - `IsConnected(ctx context.Context)`: Health check
  - `Close()`: Graceful shutdown

### Rate Limiter Component
- **File:** `backend/internal/delivery/http/middleware/rate_limiter.go`
- **Purpose:** Implements rate limiting middleware for HTTP endpoints
- **Interfaces:**
  - `NewRateLimiter(redis *infrastructure.RedisClient)`: Factory function
  - `RateLimitMiddleware(endpoint string)`: Creates middleware for specific endpoint
  - `IsRedisAvailable(ctx context.Context)`: Circuit breaker state check

### Health Check Component
- **File:** `backend/internal/delivery/http/handler.go`
- **Purpose:** Provides health monitoring endpoints
- **Interfaces:**
  - `Health(w http.ResponseWriter, r *http.Request)`: Liveness probe
  - `HealthReady(w http.ResponseWriter, r *http.Request)`: Readiness probe with Redis check

## Data Models

### Redis Key Schema
```
rate_limit:login:{ip}              → count (TTL: 60s, max: 5)
rate_limit:register:{ip}           → count (TTL: 60s, max: 3)
rate_limit:check_in:{user_id}      → count (TTL: 86400s, max: 60)
rate_limit:check_out:{user_id}     → count (TTL: 86400s, max: 60)
rate_limit:default:{ip}            → count (TTL: 60s, max: 100)

circuit_breaker:redis             → state (connected/disconnected)
circuit_breaker:last_check        → timestamp (Unix seconds)
```

### Configuration Structures
```go
type RateLimitConfig struct {
	Endpoint string
	Limit    int64
	Window   time.Duration
}

type RateLimiter struct {
	redis               *infrastructure.RedisClient
	configs             map[string]RateLimitConfig
	failOpen            bool
	circuitBreakerState string
	lastCheck           time.Time
}
```

## Technical Implementation

### 1. Redis Client Initialization (Backend)

File: `backend/internal/infrastructure/redis_client.go` (new)

```go
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

func (rc *RedisClient) IsConnected(ctx context.Context) bool {
	if err := rc.client.Ping(ctx).Err(); err != nil {
		return false
	}
	return true
}

func (rc *RedisClient) Close() error {
	log.Println("Closing Redis connection...")
	return rc.client.Close()
}

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

func (rc *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return rc.client.Get(ctx, key).Result()
}

func (rc *RedisClient) Delete(ctx context.Context, keys ...string) error {
	return rc.client.Del(ctx, keys...).Err()
}

func (rc *RedisClient) GetClient() *redis.Client {
	return rc.client
}
```

### 2. Rate Limiter Service

File: `backend/internal/delivery/http/middleware/rate_limiter.go` (new)

```go
package middleware

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/PresensiGo/backend/internal/infrastructure"
	"github.com/redis/go-redis/v9"
)

type RateLimitConfig struct {
	Endpoint string
	Limit    int64
	Window   time.Duration
}

type RateLimiter struct {
	redis   *infrastructure.RedisClient
	configs map[string]RateLimitConfig
	failOpen bool
	circuitBreakerState string
	lastCheck time.Time
}

func NewRateLimiter(redis *infrastructure.RedisClient) *RateLimiter {
	return &RateLimiter{
		redis: redis,
		configs: map[string]RateLimitConfig{
			"login":   {Endpoint: "login", Limit: 5, Window: 60 * time.Second},
			"register": {Endpoint: "register", Limit: 3, Window: 60 * time.Second},
			"check_in": {Endpoint: "check_in", Limit: 60, Window: 24 * time.Hour},
			"check_out": {Endpoint: "check_out", Limit: 60, Window: 24 * time.Hour},
			"default": {Endpoint: "default", Limit: 100, Window: 60 * time.Second},
		},
		failOpen: true, // Fail open: allow request if Redis unavailable
		circuitBreakerState: "connected",
		lastCheck: time.Now(),
	}
}

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
					log.Printf("Rate limit check failed (fail-open): %v, allowing request", err)
					next.ServeHTTP(w, r)
					return
				} else {
					// Fail closed
					http.Error(w, `{"error": "service unavailable"}`, http.StatusServiceUnavailable)
					return
				}
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.Limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", config.Limit - count + 1))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(config.Window).Unix()))

			// Check if limit exceeded
			if count > config.Limit {
				log.Printf("Rate limit exceeded: %s from %s (endpoint: %s, limit: %d, attempt: %d)", 
					r.RequestURI, clientIP, endpoint, config.Limit, count)
				http.Error(w, `{"error": "too many requests"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

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
		return realIP
	}

	// Fall back to RemoteAddr (remove port)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// No port present
		return r.RemoteAddr
	}

	return host
}

func (rl *RateLimiter) IsRedisAvailable(ctx context.Context) bool {
	available := rl.redis.IsConnected(ctx)
	
	// Update circuit breaker state
	if available {
		rl.circuitBreakerState = "connected"
	} else {
		if time.Since(rl.lastCheck) > 30*time.Second {
			rl.circuitBreakerState = "disconnected"
			log.Println("⚠️  Redis circuit breaker: OPEN (disconnected for >30s)")
		}
	}

	rl.lastCheck = time.Now()
	return available
}
```

### 3. Health Check Handler

File: `backend/internal/delivery/http/handler.go` (add methods)

```go
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) HealthReady(w http.ResponseWriter, r *http.Request) {
	// Check database connectivity (example: ping database)
	// For now, assume always ready if handler is accessible
	respondJSON(w, http.StatusOK, map[string]bool{"ready": true})
}
```

### 4. Main Application Startup

File: `backend/cmd/presensigo/main.go` (modify)

```go
func main() {
	// Initialize Redis
	redisClient, err := infrastructure.NewRedisClient("localhost:6379")
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(redisClient)

	// Initialize router
	r := mux.NewRouter()

	// Health endpoints (no rate limit)
	r.HandleFunc("/health", handler.Health).Methods("GET")
	r.HandleFunc("/health/ready", handler.HealthReady).Methods("GET")

	// Auth routes with rate limiting
	authGroup := r.NewRoute().Subrouter()
	authGroup.Use(rateLimiter.RateLimitMiddleware("login"))
	authGroup.HandleFunc("/api/auth/login", handler.Login).Methods("POST")

	authGroup2 := r.NewRoute().Subrouter()
	authGroup2.Use(rateLimiter.RateLimitMiddleware("register"))
	authGroup2.HandleFunc("/api/auth/register", handler.Register).Methods("POST")

	// Attendance routes with rate limiting
	attGroup := r.NewRoute().Subrouter()
	attGroup.Use(middleware.AuthMiddleware)
	attGroup.Use(rateLimiter.RateLimitMiddleware("check_in"))
	attGroup.HandleFunc("/api/attendance/check-in", handler.CheckIn).Methods("POST")

	attGroup2 := r.NewRoute().Subrouter()
	attGroup2.Use(middleware.AuthMiddleware)
	attGroup2.Use(rateLimiter.RateLimitMiddleware("check_out"))
	attGroup2.HandleFunc("/api/attendance/check-out", handler.CheckOut).Methods("POST")

	// Other protected routes with default rate limit
	protectedGroup := r.NewRoute().Subrouter()
	protectedGroup.Use(middleware.AuthMiddleware)
	protectedGroup.Use(rateLimiter.RateLimitMiddleware("default"))
	handler.RegisterRoutes(protectedGroup)

	// Start server
	log.Println("Server starting on :8080")
	http.ListenAndServe(":8080", r)
}
```

## Rate Limit Thresholds

| Endpoint | Limit | Window | Reason |
|----------|-------|--------|--------|
| `/api/auth/login` | 5 | 1 minute | Prevent brute force password attacks |
| `/api/auth/register` | 3 | 1 minute | Prevent account enumeration/spam |
| `/api/attendance/check-in` | 60 | 24 hours | Device/user can check-in max 60x per day |
| `/api/attendance/check-out` | 60 | 24 hours | Device/user can check-out max 60x per day |
| Other endpoints | 100 | 1 minute | Default rate limit for safety |

## Redis Configuration

**Docker Compose redis service:**
```yaml
redis:
  image: redis:7-alpine
  ports:
    - "6379:6379"
  command: redis-server --appendonly yes
  volumes:
    - redis_data:/data
```

**Connection pooling (go-redis defaults):**
- Max connections: 10
- Min idle: 5
- Connection timeout: 5 seconds
- Read timeout: 3 seconds
- Write timeout: 3 seconds

## Testing Strategy

- Unit test: IP extraction (with/without proxy, IPv6)
- Unit test: Rate limit calculation (increment, TTL, cleanup)
- Integration test: Redis connection, rate limit enforcement
- Integration test: HTTP 429 response when limit exceeded
- Integration test: X-RateLimit headers present in response
- Integration test: Redis unavailable → fail open (log warning, allow)
- Integration test: Health endpoints return correct status

## Security Considerations

- Rate limiting key never includes sensitive data (token, password)
- Client IP extracted safely (normalize, remove port)
- Rate limit logs don't leak request body (no passwords)
- Fail open protects availability over perfect rate limiting
- Circuit breaker prevents cascade failures when Redis down

## Correctness Properties

### Rate Limiting Properties
1. **Property 1: Rate Limit Enforcement**
   - For any endpoint with rate limit `N` per window `W`, no client should be able to make more than `N` requests in any `W` time window
   - **Validates:** Requirements 2.3, 2.4

2. **Property 2: Header Consistency**
   - When a request is allowed, X-RateLimit headers must accurately reflect remaining quota
   - `X-RateLimit-Remaining = X-RateLimit-Limit - current_count + 1`
   - **Validates:** Requirements 2.5

3. **Property 3: Redis Fault Tolerance**
   - When Redis is unavailable, the system must either:
     - Fail open: Allow requests and log warnings
     - Fail closed: Return HTTP 503 Service Unavailable
   - **Validates:** Requirements 4.1, 4.2, 4.3

4. **Property 4: Client IP Normalization**
   - Client IP must be extracted consistently regardless of proxy headers or port numbers
   - IP must not include port numbers or be duplicated across proxy hops
   - **Validates:** Requirements 3.1, 3.2

### Health Check Properties
5. **Property 5: Health Endpoint Availability**
   - Health endpoints must always be accessible without rate limiting or authentication
   - **Validates:** Requirements 5.5

6. **Property 6: Readiness Accuracy**
   - Readiness endpoint must accurately reflect Redis connectivity state
   - **Validates:** Requirements 5.3, 5.4

## Error Handling

### Redis Connection Errors
- **Connection Failure:** Log error and fail startup if Redis connection fails during initialization
- **Connection Loss:** Implement circuit breaker pattern with 30-second timeout before disabling rate limiting
- **Reconnection:** Automatic retry every 10 seconds when Redis is unavailable

### Rate Limiting Errors
- **Limit Exceeded:** Return HTTP 429 Too Many Requests with appropriate headers
- **Invalid Configuration:** Use default rate limits (100/min) when endpoint not configured
- **Key Generation Failures:** Use safe fallback mechanisms to prevent rate limiting bypass

### Client IP Extraction Errors
- **Malformed Headers:** Gracefully handle malformed X-Forwarded-For headers by falling back to RemoteAddr
- **IPv6 Support:** Normalize IPv6 addresses for consistent key generation
- **Private IP Ranges:** Handle private IPs appropriately for internal testing scenarios

### Graceful Degradation Strategies
1. **Fail Open (Default):** When Redis unavailable, log warning and allow requests
2. **Circuit Breaker:** After 30 seconds of Redis unavailability, disable rate limiting completely
3. **Configuration Fallbacks:** Use in-memory rate limiting as fallback (future enhancement)
4. **Health-Based Routing:** Load balancers can route traffic away from unhealthy instances

### Logging and Monitoring
- **Error Logging:** All errors logged with appropriate severity levels
- **Violation Logging:** Rate limit violations logged with IP, endpoint, and timestamp
- **State Transitions:** Circuit breaker state changes logged for monitoring
- **Health Metrics:** Redis connectivity metrics exposed for monitoring dashboards