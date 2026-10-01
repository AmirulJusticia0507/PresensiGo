# Task 5 Manual Testing Steps - Redis Rate Limiting

## Quick Start Commands

Open PowerShell in `c:\laragon\www\PresensiGo` and run:

```powershell
# 1. Run all automated tests
cd backend
$env:GOWORK = "off"
go test -v ./...
go vet ./...
go fmt ./...
go build -o ../bin/presensigo-test.exe ./cmd/api
```

**Expected:** All tests pass, no vet warnings, build succeeds

---

## Manual Testing Flow

### Step 1: Start Redis (Task 5.9)
```powershell
# From project root
docker-compose up -d redis

# Verify Redis is running
docker ps | findstr "redis"
```

### Step 2: Start Backend (Task 5.10)
```powershell
cd backend
$env:GOWORK = "off"
go run ./cmd/api/main.go
```

**Look for:**
```
✓ Connected to database
✓ Redis connected: localhost:6379
Server starting on port 8080
```

Leave this running and open a NEW PowerShell terminal for the next steps.

### Step 3: Test Rate Limiting (Task 5.11)

In the NEW terminal:
```powershell
# Run 6 login attempts rapidly
for ($i=1; $i -le 6; $i++) {
    Write-Host "`nRequest $i..."
    $response = Invoke-WebRequest -Uri "http://localhost:8080/api/auth/login" `
                                   -Method POST `
                                   -Headers @{"Content-Type"="application/json"} `
                                   -Body '{"username":"test","password":"test"}' `
                                   -UseBasicParsing
    Write-Host "Status: $($response.StatusCode)"
    Write-Host "X-RateLimit-Limit: $($response.Headers['X-RateLimit-Limit'])"
    Write-Host "X-RateLimit-Remaining: $($response.Headers['X-RateLimit-Remaining'])"
    Start-Sleep -Milliseconds 200
}
```

**Expected:**
- Requests 1-5: Status 401 or 200 (auth failure is OK)
- Request 6: **Status 429** (Too Many Requests)
- Headers show: X-RateLimit-Limit=5, X-RateLimit-Remaining=0

### Step 4: Test Health Endpoint (Task 5.12)
```powershell
Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing | Select-Object StatusCode, Content
```

**Expected:** StatusCode=200, Content=`{"status":"ok"}`

### Step 5: Test Health Ready - Redis Connected (Task 5.13)
```powershell
Invoke-WebRequest -Uri "http://localhost:8080/health/ready" -UseBasicParsing | Select-Object StatusCode, Content
```

**Expected:** StatusCode=200, Content includes `"ready":true`

### Step 6: Stop Redis and Test (Task 5.14)
```powershell
# Stop Redis
docker stop presensigo-redis

# Wait a moment
Start-Sleep -Seconds 2

# Test health ready (should fail)
try {
    Invoke-WebRequest -Uri "http://localhost:8080/health/ready" -UseBasicParsing
} catch {
    Write-Host "Status Code: $($_.Exception.Response.StatusCode.value__)"
    Write-Host "Expected: 503 Service Unavailable"
}
```

**Expected:** Status 503, ready=false

**Check backend logs** - you should see:
```
⚠️  Rate limit check failed (fail-open): ... allowing request
```

### Step 7: Restart Redis (Task 5.15)
```powershell
# Start Redis
docker start presensigo-redis

# Wait for it to be ready
Start-Sleep -Seconds 3

# Test health ready (should succeed now)
Invoke-WebRequest -Uri "http://localhost:8080/health/ready" -UseBasicParsing | Select-Object StatusCode, Content

# Test rate limiting works again
Write-Host "`nTesting rate limiting after Redis recovery..."
for ($i=1; $i -le 2; $i++) {
    Invoke-WebRequest -Uri "http://localhost:8080/api/auth/login" `
                      -Method POST `
                      -Headers @{"Content-Type"="application/json"} `
                      -Body '{"username":"test","password":"test"}' `
                      -UseBasicParsing | Select-Object StatusCode
}
```

**Expected:** 
- Health ready returns 200
- Rate limiting works (requests go through)
- Backend logs: "✓ Redis connected"

---

## Git Commit (After All Tests Pass)

```powershell
# Stop the backend server (Ctrl+C in that terminal)

# From project root
cd c:\laragon\www\PresensiGo

# Check status
git status

# Stage all changes
git add backend/internal/infrastructure/redis_client.go
git add backend/internal/delivery/http/middleware/rate_limiter.go
git add backend/internal/delivery/http/middleware/rate_limiter_test.go
git add backend/cmd/api/main.go
git add backend/internal/delivery/http/handler.go
git add .kiro/specs/p1-redis-rate-limiting/

# Commit
git commit -m "feat: implement Redis rate limiting for public endpoints

- Add Redis client with connection pooling and graceful shutdown
- Implement rate limiter middleware with per-endpoint limits
- Wire rate limiting to login (5/min), register (3/min), check-in/out (60/day)
- Add health check endpoints: /health and /health/ready
- Implement fail-open strategy for Redis unavailability
- Add circuit breaker pattern (30s threshold)
- Extract real client IP from X-Forwarded-For, X-Real-IP, RemoteAddr
- Include X-RateLimit headers in responses
- Add comprehensive unit tests for all functionality

Closes #P1-3: Redis Rate Limiting"

# Push (replace <branch> with your branch name)
git push origin <branch>
```

---

## Verification Checklist

Mark each as complete:

**Automated Tests:**
- [ ] `go test ./...` passes all tests
- [ ] `go vet ./...` has no issues
- [ ] `go fmt ./...` completes
- [ ] `go build` succeeds

**Manual Tests:**
- [ ] 5.9: Redis container starts
- [ ] 5.10: Backend server starts successfully
- [ ] 5.11: 6th login request returns 429
- [ ] 5.12: /health returns 200 OK
- [ ] 5.13: /health/ready returns 200 when Redis is up
- [ ] 5.14: /health/ready returns 503 when Redis is down
- [ ] 5.15: Rate limiting resumes after Redis restart

**Git:**
- [ ] All changes committed
- [ ] Commit message is descriptive
- [ ] Changes pushed to remote

---

## Quick Test Script

Save this as `quick-test.ps1` and run with `.\quick-test.ps1`:

```powershell
# Quick test script for rate limiting
Write-Host "=== Quick Rate Limiting Test ===" -ForegroundColor Cyan

# Test health
Write-Host "`n[1] Testing /health..." -ForegroundColor Yellow
$health = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing
Write-Host "Status: $($health.StatusCode) ✓" -ForegroundColor Green

# Test health/ready
Write-Host "`n[2] Testing /health/ready..." -ForegroundColor Yellow
$ready = Invoke-WebRequest -Uri "http://localhost:8080/health/ready" -UseBasicParsing
Write-Host "Status: $($ready.StatusCode) ✓" -ForegroundColor Green

# Test rate limiting
Write-Host "`n[3] Testing rate limiting (6 requests)..." -ForegroundColor Yellow
for ($i=1; $i -le 6; $i++) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/api/auth/login" `
                                       -Method POST `
                                       -Headers @{"Content-Type"="application/json"} `
                                       -Body '{"username":"test","password":"test"}' `
                                       -UseBasicParsing
        Write-Host "  Request $i`: Status $($response.StatusCode)"
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 429 -and $i -eq 6) {
            Write-Host "  Request $i`: Status 429 (RATE LIMITED) ✓" -ForegroundColor Green
        } else {
            Write-Host "  Request $i`: Status $statusCode"
        }
    }
    Start-Sleep -Milliseconds 200
}

Write-Host "`n=== Test Complete ===" -ForegroundColor Cyan
Write-Host "If request 6 returned 429, rate limiting is working! ✓" -ForegroundColor Green
```

---

## Troubleshooting

### "Connection refused" errors
- Check Redis is running: `docker ps | findstr redis`
- Check backend is running: Look for "Server starting" message
- Check ports: Ensure 8080 (backend) and 6379 (Redis) aren't in use

### Tests fail
```powershell
# Check Go installation
go version

# Update dependencies
go mod tidy

# Try running specific test
go test -v ./internal/delivery/http/middleware/ -run TestRateLimitMiddlewareExceeded
```

### Rate limiting doesn't work
- Check Redis connection in backend logs
- Verify middleware is wired in main.go
- Check X-RateLimit headers in response: `$response.Headers`

---

**Ready to test! Start with the automated tests, then run the manual flow.**
