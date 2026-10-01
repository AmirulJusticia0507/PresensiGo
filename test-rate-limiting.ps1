# Test Rate Limiting Implementation
# This script runs all tests and manual verification steps

Write-Host "=== Redis Rate Limiting Test Suite ===" -ForegroundColor Cyan
Write-Host ""

# Store original location
$originalLocation = Get-Location
Set-Location $PSScriptRoot

# Step 1: Run automated Go tests
Write-Host "[1/8] Running Go tests..." -ForegroundColor Yellow
Set-Location backend
$env:GOWORK = "off"
$testResult = go test ./... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Go tests PASSED" -ForegroundColor Green
    Write-Host $testResult
} else {
    Write-Host "✗ Go tests FAILED" -ForegroundColor Red
    Write-Host $testResult
    Set-Location $originalLocation
    exit 1
}
Write-Host ""

# Step 2: Run go vet
Write-Host "[2/8] Running go vet..." -ForegroundColor Yellow
$vetResult = go vet ./... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Go vet PASSED (no issues)" -ForegroundColor Green
} else {
    Write-Host "✗ Go vet found issues:" -ForegroundColor Red
    Write-Host $vetResult
}
Write-Host ""

# Step 3: Check go fmt
Write-Host "[3/8] Checking go fmt..." -ForegroundColor Yellow
$fmtResult = go fmt ./...
if ($fmtResult) {
    Write-Host "⚠ Go fmt made formatting changes in:" -ForegroundColor Yellow
    Write-Host $fmtResult
} else {
    Write-Host "✓ Go fmt check PASSED (no formatting needed)" -ForegroundColor Green
}
Write-Host ""

# Step 4: Build the application
Write-Host "[4/8] Building application..." -ForegroundColor Yellow
$buildResult = go build -o ../bin/presensigo-test.exe ./cmd/api 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Build PASSED (no errors)" -ForegroundColor Green
} else {
    Write-Host "✗ Build FAILED:" -ForegroundColor Red
    Write-Host $buildResult
    Set-Location $originalLocation
    exit 1
}
Write-Host ""

Set-Location $originalLocation

# Step 5: Start Redis (if not running)
Write-Host "[5/8] Starting Redis container..." -ForegroundColor Yellow
$redisRunning = docker ps --filter "name=presensigo-redis" --filter "status=running" -q
if (-not $redisRunning) {
    docker-compose up -d redis
    Write-Host "Waiting for Redis to start..." -ForegroundColor Yellow
    Start-Sleep -Seconds 3
    Write-Host "✓ Redis container started" -ForegroundColor Green
} else {
    Write-Host "✓ Redis already running" -ForegroundColor Green
}
Write-Host ""

# Step 6: Start the backend server
Write-Host "[6/8] Starting backend server for manual testing..." -ForegroundColor Yellow
Write-Host "Server will start in background. Use Ctrl+C to stop after testing." -ForegroundColor Cyan
Write-Host ""
Write-Host "Please run the following manual tests:" -ForegroundColor Cyan
Write-Host "  1. Test login rate limiting:" -ForegroundColor White
Write-Host "     curl -X POST http://localhost:8080/api/auth/login -H 'Content-Type: application/json' -d '{\"username\":\"test\",\"password\":\"test\"}'" -ForegroundColor Gray
Write-Host "     (Run 6 times rapidly - 6th should return 429)" -ForegroundColor Gray
Write-Host ""
Write-Host "  2. Test health endpoint:" -ForegroundColor White
Write-Host "     curl http://localhost:8080/health" -ForegroundColor Gray
Write-Host "     (Should return: {\"status\":\"ok\"})" -ForegroundColor Gray
Write-Host ""
Write-Host "  3. Test health/ready endpoint:" -ForegroundColor White
Write-Host "     curl http://localhost:8080/health/ready" -ForegroundColor Gray
Write-Host "     (Should return: {\"ready\":true})" -ForegroundColor Gray
Write-Host ""
Write-Host "  4. Stop Redis and test fail-open behavior:" -ForegroundColor White
Write-Host "     docker stop presensigo-redis" -ForegroundColor Gray
Write-Host "     curl http://localhost:8080/health/ready" -ForegroundColor Gray
Write-Host "     (Should return 503 with {\"ready\":false})" -ForegroundColor Gray
Write-Host ""
Write-Host "  5. Restart Redis and verify recovery:" -ForegroundColor White
Write-Host "     docker start presensigo-redis" -ForegroundColor Gray
Write-Host "     curl http://localhost:8080/health/ready" -ForegroundColor Gray
Write-Host "     (Should return 200 with {\"ready\":true})" -ForegroundColor Gray
Write-Host ""
Write-Host "Press Enter to start the server, then run manual tests in another terminal..." -ForegroundColor Yellow
Read-Host

Set-Location backend
$env:GOWORK = "off"
go run cmd/api/main.go

Set-Location $originalLocation
