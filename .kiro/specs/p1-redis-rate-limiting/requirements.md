# P1 #3: Activate Redis Rate Limiting

## Overview
Activate Redis rate limiting to protect against brute force attacks and DDoS. Initialize Redis connection with proper lifecycle management, wire rate-limit middleware to protected endpoints, normalize client IP identification, define per-endpoint rate limits, handle Redis unavailability gracefully, and add health check endpoint.

## Requirements

### 1. Redis Connection Lifecycle
1.1 Initialize Redis client connection on application startup in `main.go`
1.2 Implement graceful shutdown: drain pending operations, close connection
1.3 Add connection timeout (e.g., 5 seconds) and retry logic
1.4 Log connection success/failure to stdout
1.5 Verify Redis connectivity before accepting traffic (readiness check)

### 2. Rate Limit Middleware Activation
2.1 Wire rate-limit middleware to public endpoints (login, check-in, check-out, register)
2.2 Skip rate limiting for health check and internal endpoints
2.3 Apply per-endpoint limits:
   - Login: 5 attempts per minute per IP
   - Register: 3 attempts per minute per IP
   - Check-in/Check-out: 60 attempts per day per device (or per user)
   - Other endpoints: 100 requests per minute per IP (default)
2.4 Return HTTP 429 (Too Many Requests) when limit exceeded
2.5 Include rate limit headers in response (X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset)

### 3. Client IP Identification
3.1 Extract real client IP from request:
   - Remove port number from RemoteAddr
   - Check X-Forwarded-For header (for proxy scenarios)
   - Normalize IPv6 addresses if present
3.2 Use normalized IP as rate limit key (not IP:port)
3.3 Handle private IP ranges (127.0.0.1, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)

### 4. Redis Unavailability Handling
4.1 Define graceful degradation policy:
   - Option A (Fail Open): Log warning, allow request if Redis unavailable
   - Option B (Fail Closed): Deny request (return 503) if Redis unavailable
4.2 Choose Option A for MVP (fail open, prioritize availability)
4.3 Implement circuit breaker: if Redis unavailable for >30s, disable rate limiting and log alert
4.4 Retry connection every 10 seconds while unavailable
4.5 Automatic recovery: resume rate limiting when Redis comes back online

### 5. Health Check Endpoint
5.1 Add GET `/health` endpoint (public, no auth required)
5.2 Return JSON: `{"status": "ok"}` with HTTP 200
5.3 Add GET `/health/ready` endpoint (readiness check)
5.4 Readiness checks Redis and database connectivity:
   - If both connected: return `{"ready": true}` HTTP 200
   - If one or both unavailable: return `{"ready": false}` HTTP 503
5.5 No rate limiting applied to health endpoints

### 6. Monitoring & Logging
6.1 Log all rate limit violations (IP, endpoint, timestamp, limit violated)
6.2 Log Redis connection state changes (connected, disconnected, reconnected)
6.3 Expose rate limit metrics (optional): total hits, blocked requests, unique IPs
6.4 Alert if rate limiting disabled (Redis unavailable >5 min)

## Success Criteria
- Redis initialized on startup, gracefully shutdown on app exit
- Rate-limit middleware applied to login, register, check-in, check-out
- HTTP 429 returned when limit exceeded
- Real client IP extracted correctly (no port, handles proxies)
- Redis unavailability doesn't crash app (fail open)
- Health endpoints return correct status
- Rate limit headers included in responses
- No token/password leaks in rate limit logs
