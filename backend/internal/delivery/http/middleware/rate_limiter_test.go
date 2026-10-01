package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PresensiGo/backend/internal/infrastructure"
	"github.com/alicebob/miniredis/v2"
)

// TestIPExtractionWithXForwardedFor tests IP extraction with X-Forwarded-For header
func TestIPExtractionWithXForwardedFor(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := infrastructure.NewRedisClientForTesting(mr.Addr())
	limiter := NewRateLimiter(rdb)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100, 10.0.0.1, 172.16.0.1")

	ip := limiter.extractClientIP(req)
	if ip != "192.168.1.100" {
		t.Errorf("Expected 192.168.1.100, got %s", ip)
	}
}

// TestIPExtractionWithXRealIP tests IP extraction with X-Real-IP header
func TestIPExtractionWithXRealIP(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := infrastructure.NewRedisClientForTesting(mr.Addr())
	limiter := NewRateLimiter(rdb)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "203.0.113.42")

	ip := limiter.extractClientIP(req)
	if ip != "203.0.113.42" {
		t.Errorf("Expected 203.0.113.42, got %s", ip)
	}
}

// TestIPExtractionWithRemoteAddr tests IP extraction with RemoteAddr (remove port)
func TestIPExtractionWithRemoteAddr(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := infrastructure.NewRedisClientForTesting(mr.Addr())
	limiter := NewRateLimiter(rdb)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:54321"

	ip := limiter.extractClientIP(req)
	if ip != "127.0.0.1" {
		t.Errorf("Expected 127.0.0.1, got %s", ip)
	}
}

// TestRateLimitIncrement tests rate limit increment and check
func TestRateLimitIncrement(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := infrastructure.NewRedisClientForTesting(mr.Addr())
	limiter := NewRateLimiter(rdb)
	config := limiter.configs["login"]

	ctx := context.Background()
	key := "rate_limit:login:192.168.1.100"

	// First request
	count1, err := limiter.checkRateLimit(ctx, key, config)
	if err != nil {
		t.Fatalf("First check failed: %v", err)
	}
	if count1 != 1 {
		t.Errorf("Expected count 1, got %d", count1)
	}

	// Second request
	count2, err := limiter.checkRateLimit(ctx, key, config)
	if err != nil {
		t.Fatalf("Second check failed: %v", err)
	}
	if count2 != 2 {
		t.Errorf("Expected count 2, got %d", count2)
	}

	// Third request
	count3, err := limiter.checkRateLimit(ctx, key, config)
	if err != nil {
		t.Fatalf("Third check failed: %v", err)
	}
	if count3 != 3 {
		t.Errorf("Expected count 3, got %d", count3)
	}
}

// TestRateLimitMiddlewareExceeded tests HTTP 429 when limit exceeded
func TestRateLimitMiddlewareExceeded(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := infrastructure.NewRedisClientForTesting(mr.Addr())
	limiter := NewRateLimiter(rdb)

	// Create a test handler that tracks calls
	callCount := 0
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	// Wrap with rate limiter
	middleware := limiter.RateLimitMiddleware("login")
	wrappedHandler := middleware(testHandler)

	// Make requests equal to limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/api/auth/login", nil)
		req.RemoteAddr = "192.168.1.100:12345"

		recorder := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("Request %d: Expected 200, got %d", i+1, recorder.Code)
		}
	}

	// 6th request should be 429
	req := httptest.NewRequest("POST", "/api/auth/login", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	recorder := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429, got %d", recorder.Code)
	}

	// Handler should only be called 5 times (before 429)
	if callCount != 5 {
		t.Errorf("Expected handler to be called 5 times, got %d", callCount)
	}
}

// TestRateLimitHeaders tests X-RateLimit headers present in response
func TestRateLimitHeaders(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := infrastructure.NewRedisClientForTesting(mr.Addr())
	limiter := NewRateLimiter(rdb)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := limiter.RateLimitMiddleware("login")
	wrappedHandler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	recorder := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(recorder, req)

	// Check X-RateLimit headers
	if recorder.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("X-RateLimit-Limit header missing")
	}
	if recorder.Header().Get("X-RateLimit-Remaining") == "" {
		t.Error("X-RateLimit-Remaining header missing")
	}
	if recorder.Header().Get("X-RateLimit-Reset") == "" {
		t.Error("X-RateLimit-Reset header missing")
	}

	// Verify header values
	limit := recorder.Header().Get("X-RateLimit-Limit")
	if limit != "5" {
		t.Errorf("Expected limit 5, got %s", limit)
	}

	remaining := recorder.Header().Get("X-RateLimit-Remaining")
	if remaining != "4" {
		t.Errorf("Expected remaining 4, got %s", remaining)
	}
}

// TestRedisUnavailableFailOpen tests Redis unavailable fail-open behavior
func TestRedisUnavailableFailOpen(t *testing.T) {
	// Create a mock Redis client that always fails
	mockRedis := &mockRedisClient{shouldFail: true}
	limiter := NewRateLimiter(mockRedis)

	callCount := 0
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	middleware := limiter.RateLimitMiddleware("login")
	wrappedHandler := middleware(testHandler)

	req := httptest.NewRequest("POST", "/api/auth/login", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	recorder := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(recorder, req)

	// Should return 200 (fail open) even with Redis error
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected 200 (fail open), got %d", recorder.Code)
	}

	// Handler should still be called
	if callCount != 1 {
		t.Errorf("Expected handler to be called, but callCount is %d", callCount)
	}
}

// TestCircuitBreakerState tests circuit breaker state management
func TestCircuitBreakerState(t *testing.T) {
	mockRedis := &mockRedisClient{shouldFail: false}
	limiter := NewRateLimiter(mockRedis)

	ctx := context.Background()

	// Initially connected
	state := limiter.GetCircuitBreakerState()
	if state != "connected" {
		t.Errorf("Expected state 'connected', got '%s'", state)
	}

	// Check when connected
	limiter.IsRedisAvailable(ctx)
	state = limiter.GetCircuitBreakerState()
	if state != "connected" {
		t.Errorf("Expected state 'connected' after check, got '%s'", state)
	}

	// Simulate disconnection by manipulating lastCheck and failing
	mockRedis.shouldFail = true
	limiter.lastCheck = time.Now().Add(-31 * time.Second) // More than 30s ago

	limiter.IsRedisAvailable(ctx)
	state = limiter.GetCircuitBreakerState()
	if state != "disconnected" {
		t.Errorf("Expected state 'disconnected', got '%s'", state)
	}
}

// TestRateLimitPerIP tests different IPs have separate rate limits
func TestRateLimitPerIP(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := infrastructure.NewRedisClientForTesting(mr.Addr())
	limiter := NewRateLimiter(rdb)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := limiter.RateLimitMiddleware("login")
	wrappedHandler := middleware(testHandler)

	// First IP makes 5 requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/api/auth/login", nil)
		req.RemoteAddr = "192.168.1.100:12345"

		recorder := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("IP1 request %d: Expected 200, got %d", i+1, recorder.Code)
		}
	}

	// Second IP should also be able to make a request (separate limit)
	req := httptest.NewRequest("POST", "/api/auth/login", nil)
	req.RemoteAddr = "203.0.113.42:54321"

	recorder := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("IP2 request: Expected 200, got %d", recorder.Code)
	}
}

// Mock Redis client for testing fail scenarios
type mockRedisClient struct {
	shouldFail bool
}

func (m *mockRedisClient) IsConnected(ctx context.Context) bool {
	return !m.shouldFail
}

func (m *mockRedisClient) Close() error {
	return nil
}

func (m *mockRedisClient) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if m.shouldFail {
		return 0, fmt.Errorf("mock redis error")
	}
	return 1, nil
}

func (m *mockRedisClient) Get(ctx context.Context, key string) (string, error) {
	if m.shouldFail {
		return "", fmt.Errorf("mock redis error")
	}
	return "", nil
}

func (m *mockRedisClient) Delete(ctx context.Context, keys ...string) error {
	if m.shouldFail {
		return fmt.Errorf("mock redis error")
	}
	return nil
}
