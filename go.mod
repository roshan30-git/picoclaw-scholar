module github.com/user/studyclaw

go 1.21

require (
	go.mau.fi/whatsmeow v0.0.0-20240101000000-000000000000         // WhatsApp bridge
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1     // Telegram Bot API (with Mini App / WebApp support)
	github.com/pdfcpu/pdfcpu v0.6.0                                // PDF text extraction
	github.com/mattn/go-sqlite3 v1.14.22                           // SQLite driver (also used for Go Drive token store)
	github.com/robfig/cron/v3 v3.0.1                               // Job scheduler
	google.golang.org/genai v0.0.0-20240101000000-000000000000     // Gemini Flash API
	google.golang.org/api v0.170.0                                  // Google Drive API v3
	golang.org/x/oauth2 v0.18.0                                    // OAuth2 for Google Drive auth
)
