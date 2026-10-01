# Implementation Plan: Redis Rate Limiting

## Overview
Implement Redis rate limiting to protect against brute force and DDoS attacks on public endpoints (login, register, check-in, check-out). Initialize Redis connection, wire rate-limit middleware, normalize client IP, define per-endpoint limits, handle Redis unavailability, and add health check endpoint. 5 implementation tasks.

## Implementation Plan

### Phase 1: Infrastructure (Task 1-2)
- Initialize Redis client with connection pooling
- Implement rate limiter service with middleware

### Phase 2: Integration (Task 3-4)
- Wire middleware to endpoints with per-limit configuration
- Add health check endpoints

### Phase 3: Testing & Deployment (Task 5)
- Write tests, verify build, commit and push

### Estimated Timeline
- Phase 1: 1 day
- Phase 2: 1 day
- Phase 3: 1 day
- **Total: 3 days**

## Task Dependency Graph

```json
{
  "waves": [
    { "wave": 1, "tasks": ["1"] },
    { "wave": 2, "tasks": ["2"] },
    { "wave": 3, "tasks": ["3", "4"] },
    { "wave": 4, "tasks": ["5"] }
  ]
}
```

**Wave Explanation:**
- **Wave 1:** Task 1 (Redis client) independent
- **Wave 2:** Task 2 (Rate limiter) depends on Task 1
- **Wave 3:** Task 3 (Middleware wiring) and Task 4 (Health endpoints) depend on Task 2, can run parallel
- **Wave 4:** Task 5 (Testing & Commit) depends on all prior tasks

## Tasks

- [x] 1. Initialize Redis client with connection pooling
  - Create `backend/internal/infrastructure/redis_client.go`
  - Implement `NewRedisClient(addr)` with connection pooling (max: 10, min idle: 5)
  - Add `Ping()` test on initialization
  - Implement graceful shutdown: `Close()` method
  - Add timeout (5s) and retry logic
  - Log connection success/failure
  - **Files:** `backend/internal/infrastructure/redis_client.go` (new)
  - **Acceptance Criteria:**
    - Redis client connects to localhost:6379
    - Connection pooling configured (10 max, 5 min idle)
    - Graceful shutdown closes connection cleanly
    - Timeout and retry logic works
    - Connection logged to stdout

- [x] 2. Implement rate limiter middleware service
  - Create `backend/internal/delivery/http/middleware/rate_limiter.go`
  - Implement `RateLimiter` struct with Redis client
  - Implement `RateLimitMiddleware(endpoint)` function
  - Define rate limits: login 5/min, register 3/min, check-in/out 60/day, default 100/min
  - Implement `checkRateLimit()`: increment Redis counter, check against limit
  - Implement `extractClientIP()`: handle X-Forwarded-For, X-Real-IP, RemoteAddr (remove port)
  - Add rate limit headers: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
  - Implement circuit breaker for Redis unavailability (fail open, 30s threshold)
  - **Files:** `backend/internal/delivery/http/middleware/rate_limiter.go` (new)
  - **Acceptance Criteria:**
    - Rate limiter increments Redis counter per IP/endpoint
    - Returns HTTP 429 when limit exceeded
    - X-RateLimit headers present in response
    - IP extraction handles proxies (X-Forwarded-For, no port)
    - Redis unavailable → log warning, allow request (fail open)
    - Circuit breaker disables rate limiting after 30s Redis downtime

- [x] 3. Wire rate limiter middleware to endpoints
  - Modify `backend/cmd/presensigo/main.go` (or handler registration)
  - Initialize Redis client on startup
  - Initialize rate limiter with Redis client
  - Wire middleware to login route: `rateLimiter.RateLimitMiddleware("login")`
  - Wire middleware to register route: `rateLimiter.RateLimitMiddleware("register")`
  - Wire middleware to check-in route: `rateLimiter.RateLimitMiddleware("check_in")`
  - Wire middleware to check-out route: `rateLimiter.RateLimitMiddleware("check_out")`
  - Skip rate limiting on health endpoints
  - Apply default rate limiting to other protected endpoints
  - **Files:** `backend/cmd/presensigo/main.go` (modified)
  - **Acceptance Criteria:**
    - Redis client initialized on app startup
    - Rate limiter wired to login, register, check-in, check-out
    - Health endpoints not rate-limited
    - App starts with Redis connected
    - Rate limits enforced per endpoint

- [x] 4. Add health check endpoints
  - Add `GET /health` endpoint: returns `{"status": "ok"}` HTTP 200
  - Add `GET /health/ready` endpoint: returns `{"ready": true/false}` HTTP 200/503
  - Readiness check verifies Redis and database connectivity
  - Both endpoints public (no auth, no rate limit)
  - Log health check requests (optional)
  - **Files:** `backend/internal/delivery/http/handler.go` (modified)
  - **Acceptance Criteria:**
    - GET /health returns 200 with `{"status": "ok"}`
    - GET /health/ready returns 200 if Redis + DB connected
    - GET /health/ready returns 503 if Redis or DB unavailable
    - No rate limiting on health endpoints
    - Endpoints are public (accessible without auth)

- [x] 5. Write tests, verify build, and commit
  - Create `backend/internal/delivery/http/middleware/rate_limiter_test.go`
    - [x] 5.1 Test IP extraction with X-Forwarded-For header
    - [x] 5.2 Test IP extraction with X-Real-IP header
    - [x] 5.3 Test IP extraction with RemoteAddr (remove port)
    - [x] 5.4 Test rate limit increment and check
    - [x] 5.5 Test HTTP 429 when limit exceeded
    - [x] 5.6 Test X-RateLimit headers present in response
    - [x] 5.7 Test Redis unavailable scenario (fail open)
    - [x] 5.8 Test circuit breaker state management
  - Run `go test ./...` from `backend/` directory
  - Run `go vet ./...` and `go fmt` check
  - Build: `go build ./cmd/presensigo` (no errors)
  - Manual test flow:
    - [x] 5.9 Start Docker with redis: `docker compose up redis`
    - [x] 5.10 Start backend: `go run ./cmd/presensigo/main.go`
    - [x] 5.11 Login 6x in 1 minute → 6th attempt returns 429
    - [x] 5.12 Check /health → returns 200 ok
    - [x] 5.13 Check /health/ready → returns 200 ready (Redis connected)
    - [x] 5.14 Stop Redis, check /health/ready → returns 503 (Redis unavailable)
    - [x] 5.15 Stop and restart Redis → rate limiting resumes
  - Commit changes:
    ```bash
    git add backend/internal/infrastructure/redis_client.go
    git add backend/internal/delivery/http/middleware/rate_limiter.go
    git add backend/internal/delivery/http/middleware/rate_limiter_test.go
    git add backend/cmd/presensigo/main.go
    git add backend/internal/delivery/http/handler.go
    git add .kiro/specs/p1-redis-rate-limiting/
    git commit -m "feat: implement Redis rate limiting for public endpoints"
    git push
    ```
  - **Acceptance Criteria:**
    - `go test ./...` passes all tests
    - `go vet ./...` passes (no issues)
    - `go build ./cmd/presensigo` succeeds
    - Manual test flow passes all steps
    - All files committed to branch
    - Commit message is clear and descriptive

## Notes

### Rate Limit Strategy
- Login/Register: aggressive limits (5/3 per min) to prevent brute force
- Check-in/Check-out: generous daily limits (60/day) to allow legitimate usage
- Default: moderate limit (100/min) for other endpoints
- Can be adjusted per deployment environment (dev, staging, prod)

### Redis Deployment
- Docker Compose includes redis:7-alpine service
- Persistence enabled (--appendonly yes) for production
- Port 6379 exposed (default)
- Data volume mounted for durability

### Fail Open vs Fail Closed
- MVP uses fail-open: if Redis unavailable, allow requests (log warning)
- This prioritizes availability over perfect rate limiting
- Can be changed to fail-closed (deny all, return 503) if stricter requirements needed

### Client IP Identification
- Always strip port from RemoteAddr (clients behind NAT have varying ports)
- Check X-Forwarded-For header (original client IP behind proxy)
- Check X-Real-IP header (common nginx header)
- Assumes trusted proxy headers in private network or with reverse proxy validation

### Monitoring
- All violations logged with IP, endpoint, limit, timestamp
- Redis connection state changes logged
- Circuit breaker state transitions logged
- Can integrate with centralized logging/alerting (ELK, DataDog, etc.)

### Known Limitations
- Per-IP rate limits don't account for IPv6 shared addresses
- Per-user rate limits require authenticated request (not applicable to login endpoint)
- Device-based limits require device UUID binding (separate concern, P1 #1)
- No burst allowance (strict per-window limit)
- Redis cluster not configured (single instance only)
