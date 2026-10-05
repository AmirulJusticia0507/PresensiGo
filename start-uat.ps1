# PresensiGo UAT Quick Start Script
# Run from project root: .\start-uat.ps1

param(
    [string]$IP = "auto"
)

Write-Host "=== PresensiGo UAT Quick Start ===" -ForegroundColor Cyan

# Auto-detect IP if not specified
if ($IP -eq "auto") {
    $IP = (Get-NetIPAddress -AddressFamily IPv4 | Where-Object { $_.IPAddress -notlike "127.*" -and $_.IPAddress -notlike "169.*" } | Select-Object -First 1).IPAddress
    Write-Host "Auto-detected IP: $IP" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Step 1: Starting Docker services..." -ForegroundColor Green
docker compose up -d postgres redis minio

Write-Host ""
Write-Host "Step 2: Waiting for services to be healthy (30s)..." -ForegroundColor Green
Start-Sleep -Seconds 30

Write-Host ""
Write-Host "Step 3: Starting backend (new window)..." -ForegroundColor Green
Start-Process powershell -ArgumentList "-NoExit", "-Command", "Set-Location '$PSScriptRoot\backend'; go run cmd/api/main.go"

Start-Sleep -Seconds 5

Write-Host ""
Write-Host "Step 4: Seeding database..." -ForegroundColor Green
Set-Location "$PSScriptRoot\backend"
go run cmd/seed/main.go
Set-Location $PSScriptRoot

Write-Host ""
Write-Host "Step 5: Building APK with IP=$IP ..." -ForegroundColor Green
Set-Location "$PSScriptRoot\presensigo_mobile"
flutter build apk --debug --dart-define="API_BASE_URL=http://$IP`:8088/api"
Set-Location $PSScriptRoot

Write-Host ""
Write-Host "=== UAT Ready! ===" -ForegroundColor Cyan
Write-Host "APK location: presensigo_mobile\build\app\outputs\flutter-apk\app-debug.apk" -ForegroundColor White
Write-Host "Backend URL: http://$IP`:8088" -ForegroundColor White
Write-Host "Admin credentials: admin@presensigo.local / admin123" -ForegroundColor White
Write-Host "Employee credentials: employee@presensigo.local / employee123" -ForegroundColor White
Write-Host ""
Write-Host "Follow UAT_GUIDE.md for test cases." -ForegroundColor Yellow
