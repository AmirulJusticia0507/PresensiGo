@echo off
setlocal enabledelayedexpansion
cd /d "%~dp0"
set GOWORK=off

echo Testing CORS Configuration...
echo.

go test ./internal/config -run TestLoadCORSConfig -v

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo FAILED: CORS tests failed
    exit /b 1
)

echo.
echo SUCCESS: CORS tests passed
echo.

echo Building application...
go build -o api.exe ./cmd/api

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo FAILED: Build failed
    exit /b 1
)

echo.
echo SUCCESS: Application built successfully
