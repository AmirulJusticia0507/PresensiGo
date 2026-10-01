# Simple Test Runner for Rate Limiting
Write-Host "Running automated tests..." -ForegroundColor Cyan

Set-Location $PSScriptRoot\backend
$env:GOWORK = "off"

Write-Host "`n[1/3] Running go test..." -ForegroundColor Yellow
go test -v ./...
$testExit = $LASTEXITCODE

Write-Host "`n[2/3] Running go vet..." -ForegroundColor Yellow
go vet ./...
$vetExit = $LASTEXITCODE

Write-Host "`n[3/3] Running go fmt check..." -ForegroundColor Yellow
$fmtFiles = go fmt ./...

Write-Host "`n=== Summary ===" -ForegroundColor Cyan
if ($testExit -eq 0) {
    Write-Host "✓ Tests: PASSED" -ForegroundColor Green
} else {
    Write-Host "✗ Tests: FAILED" -ForegroundColor Red
}

if ($vetExit -eq 0) {
    Write-Host "✓ Vet: PASSED" -ForegroundColor Green
} else {
    Write-Host "✗ Vet: FAILED" -ForegroundColor Red
}

if (-not $fmtFiles) {
    Write-Host "✓ Fmt: PASSED" -ForegroundColor Green
} else {
    Write-Host "⚠ Fmt: Formatted files: $fmtFiles" -ForegroundColor Yellow
}

Set-Location $PSScriptRoot

if ($testExit -eq 0 -and $vetExit -eq 0) {
    Write-Host "`n✓ All checks passed!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`n✗ Some checks failed" -ForegroundColor Red
    exit 1
}
