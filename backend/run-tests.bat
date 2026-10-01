@echo off
echo Running Go tests...
cd /d "%~dp0"
set GOWORK=off

echo.
echo [1/4] Running go test...
go test -v ./...
if %ERRORLEVEL% NEQ 0 (
    echo FAILED: Tests failed
    exit /b 1
)

echo.
echo [2/4] Running go vet...
go vet ./...
if %ERRORLEVEL% NEQ 0 (
    echo WARNING: Go vet found issues
)

echo.
echo [3/4] Running go fmt...
go fmt ./...

echo.
echo [4/4] Building application...
go build -o ../bin/presensigo-test.exe ./cmd/api
if %ERRORLEVEL% NEQ 0 (
    echo FAILED: Build failed
    exit /b 1
)

echo.
echo ================================
echo All automated tests PASSED!
echo ================================
