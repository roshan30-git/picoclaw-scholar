package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

var (
	oauthConfig *oauth2.Config
	oauthState  = "studyclaw-setup-state"
	envData     map[string]string
)

const setupHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>StudyClaw Setup</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f4f4f9; display: flex; justify-content: center; align-items: center; min-height: 100vh; margin: 0; }
        .card { background: white; padding: 30px; border-radius: 12px; box-shadow: 0 4px 12px rgba(0,0,0,0.1); width: 100%; max-width: 500px; }
        h1 { margin-top: 0; color: #333; text-align: center; }
        .form-group { margin-bottom: 20px; }
        label { display: block; margin-bottom: 8px; font-weight: 500; color: #555; }
        input[type="text"], input[type="password"] { width: 100%; padding: 10px; border: 1px solid #ccc; border-radius: 6px; box-sizing: border-box; }
        textarea { width: 100%; padding: 10px; border: 1px solid #ccc; border-radius: 6px; box-sizing: border-box; resize: vertical; min-height: 80px; }
        .hint { font-size: 12px; color: #888; margin-top: 4px; }
        button { width: 100%; padding: 12px; background: #FF69B4; color: white; border: none; border-radius: 6px; font-size: 16px; font-weight: 600; cursor: pointer; transition: background 0.2s; }
        button:hover { background: #E05CA0; }
        .optional { color: #888; font-weight: normal; font-size: 13px; }
    </style>
</head>
<body>
    <div class="card">
        <h1>🦞 StudyClaw Setup</h1>
        <form action="/save" method="POST">
            <div class="form-group">
                <label>Telegram Bot Token</label>
                <input type="text" name="TELEGRAM_BOT_TOKEN" placeholder="123456789:ABCdefGHIjklMNOpqrSTUvwxYZ" required>
                <div class="hint">Get this from @BotFather on Telegram.</div>
            </div>
            <div class="form-group">
                <label>Your Phone Number</label>
                <input type="text" name="STUDYCLAW_OWNER_NUMBER" placeholder="e.g. 14155552671" required>
                <div class="hint">With country code. Used for admin commands.</div>
            </div>
            <div class="form-group">
                <label>Gemini API Key</label>
                <input type="password" name="GEMINI_API_KEY" placeholder="AIzaSy...">
                <div class="hint">Get a free key from Google AI Studio.</div>
            </div>
            <div class="form-group">
                <label>OpenAI API Key <span class="optional">(Optional)</span></label>
                <input type="password" name="OPENAI_API_KEY" placeholder="sk-...">
            </div>
            <div class="form-group">
                <label>Google Drive/Classroom Credentials JSON <span class="optional">(Optional)</span></label>
                <textarea name="GOOGLE_CREDENTIALS" placeholder="{&quot;installed&quot;:{...}}"></textarea>
                <div class="hint">Paste your Google Cloud OAuth 2.0 Client ID JSON here to enable GDrive integration. If provided, you will be redirected to log in with Google.</div>
            </div>
            <button type="submit">Complete Setup</button>
        </form>
    </div>
</body>
</html>
`

const successHTML = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8"><title>Setup Complete</title>
    <style>body { font-family: sans-serif; text-align: center; padding: 50px; }</style>
</head>
<body>
    <h1>✅ Setup Complete!</h1>
    <p>You can close this tab and return to the terminal.</p>
    <p>StudyClaw will now continue launching.</p>
</body>
</html>
`

func writeEnv(env map[string]string) error {
	f, err := os.OpenFile(".env", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	for k, v := range env {
		if v != "" {
			_, _ = f.WriteString(fmt.Sprintf("%s=%s\n", k, v))
		}
	}
	return nil
}

func saveGoogleCredentials(jsonContent string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".studyclaw")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, "google_credentials.json")
	return os.WriteFile(path, []byte(jsonContent), 0644)
}

func saveGoogleToken(token *oauth2.Token) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".studyclaw", "google_token.json")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

// RunServerIfConfigMissing checks if critical config is missing.
// If it is, it starts a blocking web server on port 8080 until setup is complete.
func RunServerIfConfigMissing() {
	_ = godotenv.Load()

	gemini := os.Getenv("GEMINI_API_KEY")
	openai := os.Getenv("OPENAI_API_KEY")
	tg := os.Getenv("TELEGRAM_BOT_TOKEN")

	if (gemini != "" || openai != "") && tg != "" {
		return // Initialized
	}

	fmt.Println("======================================================")
	fmt.Println("🚀 Initial Setup Required!")
	fmt.Println("Please open your browser and visit:")
	fmt.Println("👉 http://localhost:8080/setup")
	fmt.Println("======================================================")

	mux := http.NewServeMux()
	server := &http.Server{Addr: ":8080", Handler: mux}

	done := make(chan bool)

	mux.HandleFunc("/setup", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(setupHTML))
	})

	mux.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/setup", http.StatusSeeOther)
			return
		}

		envData = map[string]string{
			"TELEGRAM_BOT_TOKEN":     r.FormValue("TELEGRAM_BOT_TOKEN"),
			"STUDYCLAW_OWNER_NUMBER": r.FormValue("STUDYCLAW_OWNER_NUMBER"),
			"GEMINI_API_KEY":         r.FormValue("GEMINI_API_KEY"),
			"OPENAI_API_KEY":         r.FormValue("OPENAI_API_KEY"),
		}

		googleCreds := strings.TrimSpace(r.FormValue("GOOGLE_CREDENTIALS"))
		if googleCreds != "" {
			if err := saveGoogleCredentials(googleCreds); err != nil {
				http.Error(w, "Failed to save Google credentials: "+err.Error(), 500)
				return
			}

			// Enable GDrive implicitly
			envData["STUDYCLAW_ENABLE_GDRIVE"] = "true"

			cfg, err := google.ConfigFromJSON([]byte(googleCreds), drive.DriveFileScope)
			if err != nil {
				http.Error(w, "Invalid Google JSON: "+err.Error(), 400)
				return
			}
			cfg.RedirectURL = "http://localhost:8080/auth/google/callback"
			oauthConfig = cfg

			authURL := oauthConfig.AuthCodeURL(oauthState, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
			http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
			return
		}

		if err := writeEnv(envData); err != nil {
			http.Error(w, "Failed to write .env", 500)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(successHTML))
		go func() {
			time.Sleep(1 * time.Second)
			server.Shutdown(context.Background())
			done <- true
		}()
	})

	mux.HandleFunc("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		state := r.FormValue("state")
		if state != oauthState {
			http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
			return
		}

		code := r.FormValue("code")
		if code == "" {
			http.Error(w, "Code not found", http.StatusBadRequest)
			return
		}

		token, err := oauthConfig.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "Token exchange failed: "+err.Error(), 500)
			return
		}

		if err := saveGoogleToken(token); err != nil {
			http.Error(w, "Failed to save token: "+err.Error(), 500)
			return
		}

		if err := writeEnv(envData); err != nil {
			http.Error(w, "Failed to write .env", 500)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(successHTML))

		go func() {
			time.Sleep(1 * time.Second)
			server.Shutdown(context.Background())
			done <- true
		}()
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Setup server failed: %v", err)
		}
	}()

	<-done
	fmt.Println("✅ Setup completed. Resuming startup...")
	_ = godotenv.Load() // Reload the newly created .env variables into the current process
}
