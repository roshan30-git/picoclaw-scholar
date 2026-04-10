#!/usr/bin/env bash
# StudyClaw - Termux Launcher & Setup Wizard
# Works on Termux (Android) with POSIX sh

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

show_header() {
    clear
    echo -e "${CYAN}StudyClaw - Termux Orchestrator${NC}"
    echo "================================"
    echo ""
}

show_header

# 0. Self-Update Check
echo -e "${CYAN}[UPDATE] Syncing with GitHub...${NC}"
if command -v git &> /dev/null; then
    git fetch origin main &> /dev/null || true
    git reset --hard origin/main &> /dev/null || true
    echo -e "${GREEN}[SUCCESS] Up to date.${NC}"
else
    echo -e "${YELLOW}[WARNING] Git not found. Skipping auto-update.${NC}"
fi
echo ""

# 1. Check for Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}[ERROR] Go is not installed.${NC}"
    echo "Install it with: pkg install golang"
    exit 1
fi

# 2. Setup .env if missing or empty GEMINI key
if [ -f .env ]; then
    GEMINI_API_KEY=$(grep "^GEMINI_API_KEY=" .env | cut -d'=' -f2 | tr -d '[:space:]')
fi

if [ -z "${GEMINI_API_KEY}" ]; then
    echo -e "${YELLOW}Welcome! Your Gemini API Key is missing.${NC}"
    echo ""
    printf "Enter your GEMINI_API_KEY (from https://aistudio.google.com/apikey): "
    read -r GEMINI_API_KEY
    while [ -z "$(echo "$GEMINI_API_KEY" | tr -d '[:space:]')" ]; do
        echo -e "${RED}   Key cannot be empty!${NC}"
        printf "Enter your GEMINI_API_KEY: "
        read -r GEMINI_API_KEY
    done

    printf "Enter your TELEGRAM_BOT_TOKEN (optional, press Enter to skip): "
    read -r TG_TOKEN

    printf "Enter your WhatsApp number (e.g. 91XXXXXXXXXX, optional): "
    read -r OWNER_NUM

    echo ""
    echo -e "${CYAN}[CONFIG] Saving configuration to .env...${NC}"

    cat > .env << EOF
GEMINI_API_KEY=${GEMINI_API_KEY}
TELEGRAM_BOT_TOKEN=${TG_TOKEN}
STUDYCLAW_OWNER_NUMBER=${OWNER_NUM}
LLM_PROVIDER=gemini
EOF

    echo -e "${GREEN}[SUCCESS] Configuration saved!${NC}"
    echo ""
fi

# 3. Synchronize & Compile (Efficient for 4GB RAM)
echo -e "${CYAN}[BUILD] Compiling StudyClaw binary...${NC}"
go build -o studyclaw ./cmd/main.go
echo -e "${GREEN}[SUCCESS] Binary ready.${NC}"

# 4. Termux Specific: Wake-lock (Prevents Android from killing the bot)
if command -v termux-wake-lock &> /dev/null; then
    echo -e "${CYAN}[TERMUX] Battery optimization: Requesting Wake-lock...${NC}"
    termux-wake-lock
fi

# 5. Launch
show_header
echo -e "${GREEN}[LAUNCH] StudyClaw is running!${NC}"
echo "   (Send !stop from WhatsApp to shut down gracefully)"
echo ""

# Cleanup on exit
trap 'if command -v termux-wake-unlock &> /dev/null; then termux-wake-unlock; fi; exit' INT TERM EXIT

./studyclaw
