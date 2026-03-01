# StudyClaw Windows Launcher & Setup Wizard
# This script ensures the environment is set up and runs the application.

$ErrorActionPreference = "Stop"

function Show-Header {
    Clear-Host
    Write-Host "🦞 " -NoNewline
    Write-Host "StudyClaw - Windows Orchestrator" -ForegroundColor Cyan
    Write-Host "================================" -ForegroundColor Gray
    Write-Host ""
}

Show-Header

# 1. Check for Go
if (!(Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Error: Go is not installed." -ForegroundColor Red
    Write-Host "Please install it from https://golang.org/dl/" -ForegroundColor White
    exit 1
}

# 2. Setup .env
if (!(Test-Path .env)) {
    Write-Host "👋 Welcome! Let's set up StudyClaw for the first time." -ForegroundColor Yellow
    Write-Host ""

    $geminiKey = ""
    while ($geminiKey -eq "") {
        $geminiKey = Read-Host "🔑 Enter your GEMINI_API_KEY (from https://aistudio.google.com/apikey)"
        if ($geminiKey -eq "") { Write-Host "   Key cannot be empty!" -ForegroundColor Red }
    }

    $ownerNumber = Read-Host "📱 Enter your WhatsApp number (e.g. 91XXXXXXXXXX, leave empty for now)"
    $tgToken = Read-Host "🤖 Enter your TELEGRAM_BOT_TOKEN (leave empty to skip)"

    Write-Host ""
    Write-Host "⚙️  Saving configuration..." -ForegroundColor Gray
    
    $envContent = @"
GEMINI_API_KEY=$($geminiKey.Trim())
TELEGRAM_BOT_TOKEN=$($tgToken.Trim())
STUDYCLAW_OWNER_NUMBER=$($ownerNumber.Trim())
LLM_PROVIDER=gemini
"@
    
    Set-Content -Path .env -Value $envContent -Encoding UTF8
    Write-Host "✅ .env created successfully!" -ForegroundColor Green
    Write-Host ""
}

# 3. Synchronize Dependencies
Write-Host "📦 Refreshing dependencies..." -ForegroundColor Gray
go mod tidy

# 4. Launch
Show-Header
Write-Host "🚀 Launching StudyClaw..." -ForegroundColor Green
Write-Host "   (Press Ctrl+C to stop)" -ForegroundColor Gray
Write-Host ""
go run ./cmd/main.go
