# StudyClaw Windows Launcher
# This script ensures the environment is set up and runs the application.

$ErrorActionPreference = "Stop"

Write-Host "StudyClaw - Windows Orchestrator" -ForegroundColor Cyan

# 1. Check for Go
if (!(Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "Error: Go is not installed. Please install it from https://golang.org/dl/" -ForegroundColor Red
    exit 1
}

# 2. Setup .env if missing
if (!(Test-Path .env)) {
    Write-Host "Setting up your .env file..." -ForegroundColor Yellow
    $geminiKey = Read-Host "Enter your GEMINI_API_KEY"
    $tgToken = Read-Host "Enter your TELEGRAM_BOT_TOKEN (leave empty to skip)"
    
    $envContent = "GEMINI_API_KEY=$geminiKey`nTELEGRAM_BOT_TOKEN=$tgToken`nLLM_PROVIDER=gemini"
    Set-Content -Path .env -Value $envContent -Encoding UTF8
    Write-Host "Done: .env created successfully." -ForegroundColor Green
}

# 3. Synchronize Dependencies
Write-Host "Refreshing dependencies..." -ForegroundColor Gray
go mod tidy

# 4. Launch
Write-Host "Launching StudyClaw..." -ForegroundColor Green
go run ./cmd/main.go
