# PresensiGo - Development Script (PowerShell)

param(
    [string]$Command = "help"
)

switch ($Command) {
    "dev" {
        Write-Host "Starting development server..." -ForegroundColor Green
        Set-Location backend
        $env:GOWORK = "off"
        go run cmd/api/main.go
    }
    "build" {
        Write-Host "Building backend..." -ForegroundColor Green
        Set-Location backend
        $env:GOWORK = "off"
        go build -o ../bin/presensigo.exe ./cmd/api
        Write-Host "Build complete: bin/presensigo.exe" -ForegroundColor Green
    }
    "db-up" {
        Write-Host "Starting database containers..." -ForegroundColor Green
        docker-compose up -d postgres redis minio
        Write-Host "Waiting for PostgreSQL..." -ForegroundColor Yellow
        Start-Sleep -Seconds 3
        Write-Host "Containers are running!" -ForegroundColor Green
    }
    "db-down" {
        Write-Host "Stopping containers..." -ForegroundColor Yellow
        docker-compose down
    }
    "db-reset" {
        Write-Host "Resetting database..." -ForegroundColor Yellow
        docker-compose down -v
        docker-compose up -d postgres redis minio
        Write-Host "Database reset complete!" -ForegroundColor Green
    }
    "db-shell" {
        docker exec -it presensigo-postgres psql -U presensigo -d presensigo
    }
    "deps" {
        Write-Host "Installing dependencies..." -ForegroundColor Green
        Set-Location backend
        $env:GOWORK = "off"
        go mod tidy
    }
    "setup" {
        & "$PSCommandPath" "db-up"
        & "$PSCommandPath" "deps"
        Write-Host "Setup complete! Run: .\dev.ps1 dev" -ForegroundColor Green
    }
    default {
        Write-Host "PresensiGo Development Commands:" -ForegroundColor Cyan
        Write-Host "  .\dev.ps1 dev       - Start development server" -ForegroundColor White
        Write-Host "  .\dev.ps1 build     - Build backend binary" -ForegroundColor White
        Write-Host "  .\dev.ps1 db-up     - Start Docker containers" -ForegroundColor White
        Write-Host "  .\dev.ps1 db-down   - Stop Docker containers" -ForegroundColor White
        Write-Host "  .\dev.ps1 db-reset  - Reset database" -ForegroundColor White
        Write-Host "  .\dev.ps1 db-shell  - Connect to PostgreSQL" -ForegroundColor White
        Write-Host "  .\dev.ps1 deps      - Install Go dependencies" -ForegroundColor White
        Write-Host "  .\dev.ps1 setup     - Full setup (db-up + deps)" -ForegroundColor White
    }
}
