# StudyClaw Windows Launcher & Setup Wizard
# This script ensures the environment is set up and runs the application.

$ErrorActionPreference = "Stop"

function Show-Header {
    Clear-Host
    Write-Host "StudyClaw - Windows Orchestrator" -ForegroundColor Cyan
    Write-Host "================================" -ForegroundColor Gray
    Write-Host ""
}

# Helper to load .env into current session
function Load-Env {
    if (Test-Path .env) {
        Get-Content .env | ForEach-Object {
            $line = $_.Trim()
            if ($line -and !$line.StartsWith("#") -and $line.Contains("=")) {
                $name, $value = $line.Split('=', 2)
                [System.Environment]::SetEnvironmentVariable($name.Trim(), $value.Trim(), "Process")
            }
        }
    }
}

Show-Header

# 1. Check for Go
if (!(Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "[ERROR] Go is not installed." -ForegroundColor Red
    Write-Host "Please install it from https://golang.org/dl/" -ForegroundColor White
    exit 1
}

# 2. Check and Setup Configuration
Load-Env
$geminiKey = [System.Environment]::GetEnvironmentVariable("GEMINI_API_KEY", "Process")

if ([string]::IsNullOrWhiteSpace($geminiKey)) {
    Write-Host "Welcome! It looks like your Gemini API Key is missing." -ForegroundColor Yellow
    Write-Host "This key is required for the AI to function."
    Write-Host ""

    $geminiKey = ""
    while ([string]::IsNullOrWhiteSpace($geminiKey)) {
        $geminiKey = Read-Host "Enter your GEMINI_API_KEY (get one at https://aistudio.google.com/apikey)"
        if ([string]::IsNullOrWhiteSpace($geminiKey)) { 
            Write-Host "   Key cannot be empty!" -ForegroundColor Red 
        }
    }

    $tgToken = [System.Environment]::GetEnvironmentVariable("TELEGRAM_BOT_TOKEN", "Process")
    if ([string]::IsNullOrWhiteSpace($tgToken)) {
        $tgToken = Read-Host "Enter your TELEGRAM_BOT_TOKEN (optional, press Enter to skip)"
    }

    $ownerNumber = [System.Environment]::GetEnvironmentVariable("STUDYCLAW_OWNER_NUMBER", "Process")
    if ([string]::IsNullOrWhiteSpace($ownerNumber)) {
        $ownerNumber = Read-Host "Enter your WhatsApp number (e.g. 91XXXXXXXXXX, optional)"
    }

    Write-Host ""
    Write-Host "[CONFIG] Saving configuration to .env..." -ForegroundColor Gray
    
    $envContent = @"
GEMINI_API_KEY=$($geminiKey.Trim())
TELEGRAM_BOT_TOKEN=$($tgToken.Trim())
STUDYCLAW_OWNER_NUMBER=$($ownerNumber.Trim())
LLM_PROVIDER=gemini
"@
    
    $envContent | Out-File -FilePath .env -Encoding utf8
    Write-Host "[SUCCESS] Configuration saved!" -ForegroundColor Green
    Write-Host ""
}

# 3. Synchronize Dependencies
Write-Host "[DEPS] Refreshing dependencies..." -ForegroundColor Gray
go mod tidy

# 4. Launch
Show-Header
Write-Host "[LAUNCH] Launching StudyClaw..." -ForegroundColor Green
Write-Host "   (Press Ctrl+C to stop)" -ForegroundColor Gray
Write-Host ""
go run ./cmd/main.go
