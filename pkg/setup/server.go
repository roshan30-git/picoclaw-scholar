package setup

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/roshan30-git/picoclaw-scholar/pkg/auth"
	"github.com/roshan30-git/picoclaw-scholar/pkg/logger"
	"golang.org/x/oauth2"
)

var (
	oauthState      = "studyclaw-setup-state"
	pendingVerifier string
	envData         map[string]string
)

const setupHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>StudyClaw | Premium Setup</title>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;800&display=swap" rel="stylesheet">
    <style>
        :root {
            --primary: #FF69B4;
            --secondary: #7000FF;
            --bg: #0a0a14;
        }
        body {
            margin: 0;
            font-family: 'Inter', sans-serif;
            background: var(--bg);
            background-image: 
                radial-gradient(at 0% 0%, rgba(112, 0, 255, 0.15) 0px, transparent 50%),
                radial-gradient(at 100% 100%, rgba(255, 105, 180, 0.15) 0px, transparent 50%);
            color: white;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            overflow-x: hidden;
        }
        .container {
            width: 100%;
            max-width: 550px;
            padding: 20px;
            z-index: 1;
        }
        .glass-card {
            background: rgba(255, 255, 255, 0.03);
            backdrop-filter: blur(20px);
            -webkit-backdrop-filter: blur(20px);
            border: 1px solid rgba(255, 255, 255, 0.05);
            border-radius: 24px;
            padding: 40px;
            box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
        }
        .logo {
            font-size: 48px;
            text-align: center;
            margin-bottom: 10px;
        }
        h1 {
            font-size: 28px;
            font-weight: 800;
            text-align: center;
            margin: 0 0 30px 0;
            background: linear-gradient(135deg, #fff 0%, #aaa 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        .section-title {
            font-size: 14px;
            text-transform: uppercase;
            letter-spacing: 1.5px;
            color: rgba(255, 255, 255, 0.4);
            margin-bottom: 20px;
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
            padding-bottom: 10px;
        }
        .form-group {
            margin-bottom: 25px;
        }
        label {
            display: block;
            margin-bottom: 10px;
            font-size: 14px;
            font-weight: 600;
            color: rgba(255, 255, 255, 0.8);
        }
        input {
            width: 100%;
            padding: 14px 18px;
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 12px;
            color: white;
            font-size: 15px;
            transition: all 0.3s ease;
            box-sizing: border-box;
        }
        input:focus {
            outline: none;
            border-color: var(--primary);
            background: rgba(255, 255, 255, 0.08);
            box-shadow: 0 0 0 4px rgba(255, 105, 180, 0.1);
        }
        .btn-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 15px;
            margin-bottom: 30px;
        }
        .connect-btn {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 10px;
            padding: 14px;
            border-radius: 12px;
            text-decoration: none;
            font-weight: 600;
            font-size: 14px;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            border: 1px solid transparent;
            cursor: pointer;
        }
        .btn-google {
            background: rgba(255, 255, 255, 0.05);
            color: white;
            border-color: rgba(255, 255, 255, 0.1);
        }
        .btn-google:hover {
            background: rgba(255, 255, 255, 0.1);
            transform: translateY(-2px);
        }
        .btn-chatgpt {
            background: linear-gradient(135deg, var(--secondary), #4a00ff);
            color: white;
            box-shadow: 0 10px 20px -5px rgba(112, 0, 255, 0.3);
        }
        .btn-chatgpt:hover {
            transform: translateY(-2px);
            box-shadow: 0 15px 25px -5px rgba(112, 0, 255, 0.4);
        }
        .submit-btn {
            width: 100%;
            padding: 16px;
            background: linear-gradient(135deg, var(--primary), #ff1493);
            color: white;
            border: none;
            border-radius: 14px;
            font-size: 17px;
            font-weight: 800;
            cursor: pointer;
            transition: all 0.3s ease;
            box-shadow: 0 10px 20px -5px rgba(255, 105, 180, 0.3);
        }
        .submit-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 15px 25px -5px rgba(255, 105, 180, 0.4);
        }
        .hint {
            font-size: 12px;
            color: rgba(255, 255, 255, 0.4);
            margin-top: 6px;
        }
        .icon { width: 20px; height: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="glass-card">
            <div class="logo">🦞</div>
            <h1>StudyClaw Setup</h1>
            
            <div class="section-title">Step 1: Connect Accounts</div>
            <div class="btn-grid">
                <a href="/auth/google" class="connect-btn btn-google">
                    <img src="https://www.google.com/favicon.ico" class="icon" alt="">
                    Google Drive
                </a>
                <a href="/auth/chatgpt" class="connect-btn btn-chatgpt">
                    <span>✨</span>
                    ChatGPT Pro
                </a>
            </div>

            <div class="section-title">Step 2: Configuration</div>
            <form action="/save" method="POST">
                <div class="form-group">
                    <label for="TELEGRAM_BOT_TOKEN">Telegram Bot Token</label>
                    <input type="text" id="TELEGRAM_BOT_TOKEN" name="TELEGRAM_BOT_TOKEN" placeholder="123456789:ABC..." aria-describedby="hint-TELEGRAM_BOT_TOKEN" required>
                    <div id="hint-TELEGRAM_BOT_TOKEN" class="hint">Get from @BotFather</div>
                </div>
                <div class="form-group">
                    <label for="STUDYCLAW_OWNER_NUMBER">Admin Phone Number</label>
                    <input type="text" id="STUDYCLAW_OWNER_NUMBER" name="STUDYCLAW_OWNER_NUMBER" placeholder="e.g. 919832XXXXXX" aria-describedby="hint-STUDYCLAW_OWNER_NUMBER" required>
                    <div id="hint-STUDYCLAW_OWNER_NUMBER" class="hint">With country code, no "+"</div>
                </div>
                <div class="form-group">
                    <label for="GEMINI_API_KEY">Gemini API Key (Optional)</label>
                    <input type="password" id="GEMINI_API_KEY" name="GEMINI_API_KEY" placeholder="AIzaSy..." aria-describedby="hint-GEMINI_API_KEY">
                    <div id="hint-GEMINI_API_KEY" class="hint">For fallback or direct vision features</div>
                </div>
                
                <button type="submit" class="submit-btn">Complete Setup & Launch 🚀</button>
            </form>
        </div>
    </div>
</body>
</html>
`

const successHTML = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8"><title>Setup Complete</title>
    <style>
        body { font-family: 'Inter', sans-serif; text-align: center; padding: 100px; background: #0a0a14; color: white; }
        .success-logo { font-size: 80px; margin-bottom: 20px; }
        h1 { font-size: 40px; margin-bottom: 10px; }
        p { color: rgba(255,255,255,0.6); font-size: 18px; }
    </style>
</head>
<body>
    <div class="success-logo">✅</div>
    <h1>Setup Complete!</h1>
    <p>StudyClaw is now fully configured and launching.</p>
    <p>You can close this tab and return to your terminal.</p>
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

func saveGoogleToken(token *oauth2.Token) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".studyclaw", "google_token.json")
	dir := filepath.Dir(path)
	_ = os.MkdirAll(dir, 0755)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

func generateState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// RunServerIfConfigMissing checks if critical config is missing.
// If it is, it starts a blocking web server on port 8080 until setup is complete.
func RunServerIfConfigMissing() {
	_ = godotenv.Load()

	tg := os.Getenv("TELEGRAM_BOT_TOKEN")
	owner := os.Getenv("STUDYCLAW_OWNER_NUMBER")

	if tg != "" && owner != "" {
		return // Initialized
	}

	logger.InfoC("setup", "======================================================")
	logger.InfoC("setup", "🚀 StudyClaw: Initial Setup Required!")
	logger.InfoC("setup", "Please open your browser and visit:")
	logger.InfoC("setup", "👉 http://localhost:8080/setup")
	logger.InfoC("setup", "======================================================")

	mux := http.NewServeMux()
	server := &http.Server{Addr: ":8080", Handler: mux}

	done := make(chan bool)

	mux.HandleFunc("/setup", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(setupHTML))
	})

	mux.HandleFunc("/auth/google", func(w http.ResponseWriter, r *http.Request) {
		cfg := auth.GoogleAntigravityOAuthConfig()
		pkce, _ := auth.GeneratePKCE()
		state, _ := generateState()

		redirectURI := "http://localhost:8080/auth/google/callback"
		authURL := auth.BuildAuthorizeURL(cfg, pkce, state, redirectURI)

		oauthState = state
		pendingVerifier = pkce.CodeVerifier

		http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	})

	mux.HandleFunc("/auth/chatgpt", func(w http.ResponseWriter, r *http.Request) {
		cfg := auth.OpenAIOAuthConfig()
		pkce, _ := auth.GeneratePKCE()
		state, _ := generateState()

		redirectURI := "http://localhost:8080/auth/chatgpt/callback"
		authURL := auth.BuildAuthorizeURL(cfg, pkce, state, redirectURI)

		oauthState = state
		pendingVerifier = pkce.CodeVerifier

		http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	})

	mux.HandleFunc("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		state := r.FormValue("state")
		if state != oauthState {
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}

		code := r.FormValue("code")
		cfg := auth.GoogleAntigravityOAuthConfig()
		redirectURI := "http://localhost:8080/auth/google/callback"

		cred, err := auth.ExchangeCodeForTokens(cfg, code, pendingVerifier, redirectURI)
		if err != nil {
			http.Error(w, "Auth failed: "+err.Error(), 500)
			return
		}

		_ = auth.SetCredential("google-antigravity", cred)

		token := &oauth2.Token{
			AccessToken:  cred.AccessToken,
			RefreshToken: cred.RefreshToken,
			Expiry:       cred.ExpiresAt,
			TokenType:    "Bearer",
		}
		_ = saveGoogleToken(token)

		http.Redirect(w, r, "/setup", http.StatusSeeOther)
	})

	mux.HandleFunc("/auth/chatgpt/callback", func(w http.ResponseWriter, r *http.Request) {
		state := r.FormValue("state")
		if state != oauthState {
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}

		code := r.FormValue("code")
		cfg := auth.OpenAIOAuthConfig()
		redirectURI := "http://localhost:8080/auth/chatgpt/callback"

		cred, err := auth.ExchangeCodeForTokens(cfg, code, pendingVerifier, redirectURI)
		if err != nil {
			http.Error(w, "Auth failed: "+err.Error(), 500)
			return
		}

		_ = auth.SetCredential("openai", cred)
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
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
		}

		cred, _ := auth.GetCredential("google-antigravity")
		if cred != nil {
			envData["STUDYCLAW_ENABLE_GDRIVE"] = "true"
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
	logger.InfoC("setup", "✅ Setup completed. Resuming startup...")
	_ = godotenv.Load()
}
