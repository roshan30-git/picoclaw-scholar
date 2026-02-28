<div align="center">
  <h1>🦞 StudyClaw</h1>
  <h3>Your autonomous AI study companion — lives in WhatsApp & Telegram, runs on your Android phone or Windows PC.</h3>
  <br/>
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white">
  <img src="https://img.shields.io/badge/Platform-Android%20%7C%20Termux-orange">
  <img src="https://img.shields.io/badge/AI-Gemini%202.0%20Flash%20Lite-blue?logo=google">
  <img src="https://img.shields.io/badge/DB-SQLite-003B57?logo=sqlite">
  <img src="https://img.shields.io/badge/RAM-~30MB-brightgreen">
  <img src="https://img.shields.io/badge/license-MIT-green">
</div>

---

> **StudyClaw** is a Go-based agentic AI bot that runs directly on your Android phone via [Termux](https://termux.dev) or natively on **Windows**. It connects to WhatsApp and Telegram, listens for PDFs and study messages, indexes them into a local SQLite database, and proactively sends you quizzes, diagrams, and summaries — powered entirely by the **Gemini free tier** (with options for ChatGPT Pro/Gemini CLI).

---

## ✨ Features

| Feature | Description |
|:--------|:------------|
| 📥 **PDF Indexing** | Send any PDF to the bot. It extracts text and stores it in a local knowledge base. |
| 🎯 **AI Quizzes** | Ask for a quiz and get 3–5 MCQs drawn from your own notes via Gemini. |
| 📊 **Diagram Viewer** | If the AI generates a Mermaid diagram, a local `http://127.0.0.1:8080` link is sent so you can view it in your browser with full pan/zoom. |
| 💬 **Multi-Channel** | Run the bot on WhatsApp (via paired device) or Telegram (via long-polling). |
| 🔑 **Bring Your Own Key** | Default free Gemini tier, but natively supports auth via ChatGPT Pro (`codex`) or Gemini CLI (`antigravity`). |
| 📅 **Calendar Engine** | Keeps track of your semester dates and exams to adjust its "soul" (Drill Sergeant vs. Peer). |
| 🤖 **Native Tool Calling** | Uses Gemini's native function calling to autonomously decide when to generate quizzes or index new documents. |

---

## 🏗️ Architecture

```
studyclaw/
├── cmd/
│   ├── main.go          ← Entry point (start here)
│   └── setup.go         ← Interactive setup wizard
├── integrations/
│   ├── whatsapp/        ← whatsmeow-based channel (Channel interface)
│   └── gdrive/          ← Google Drive OAuth2 sync
├── pkg/
│   ├── agent/           ← Agent loop (session memory + tool routing)
│   ├── bus/             ← In-process message bus (pub/sub)
│   ├── channels/        ← Channel interface + Manager + Telegram stub
│   ├── config/          ← Config struct and defaults
│   ├── database/        ← SQLite: notes, embeddings, quiz history
│   ├── memory/          ← Reflection Manager & Session History
│   ├── providers/       ← Gemini 2.0 provider (rate-limited, retry on 429)
│   ├── study/           ← Quiz engine, Ingest engine, Calendar Engine, OCR
│   ├── tools/           ← Tool interface + Report/PYQ tools
│   ├── viewer/          ← Local HTTP server + viewer.html
│   └── visual/          ← Diagram management & Circuit viewer
└── workspace/
    └── PROMPTS/         ← Persona files (base_soul, drill_sergeant, etc.)
```

---

## 🚀 Quick Start (Windows)

StudyClaw is natively designed to be completely setup-free on Windows using its unified orchestrator script. **No C++ compilers required.**

### 1. Requirements
Ensure you have the Go compiler installed from [golang.org/dl](https://golang.org/dl/).

### 2. Launch
Download the repository, open PowerShell in the folder, and run:
```powershell
.\run.ps1
```
The interactive orchestrator will detect if you are missing API keys, auto-configure your `.env` file, download dependencies, and boot the AI loop in seconds.

---

## 🚀 Quick Start (Termux on Android)
### 1. Install Termux & Environment
Download **Termux** from F-Droid. Open it and run:
```bash
pkg update && pkg upgrade -y
pkg install golang git clang make -y
```

### 2. Clone & Setup
```bash
git clone https://github.com/roshan30-git/picoclaw-scholar.git
cd picoclaw-scholar
go mod tidy
```

### 3. API Key Configuration
Get your free key at [aistudio.google.com](https://aistudio.google.com/app/apikey).
```bash
export GEMINI_API_KEY="AIzaSy..."
# Optional: Enabale Telegram Bot Support
export TELEGRAM_BOT_TOKEN="12345:ABCDE..."
# Optional: Set your owner number for admin access
export STUDYCLAW_OWNER_NUMBER="91XXXXXXXXXX"
```

### 4. Launch StudyClaw
```bash
go run cmd/main.go
```

**Next Steps:**
- A **QR Code** will print in Termux.
- Take a screenshot of it or share the Termux link to your PC.
- Open WhatsApp on your phone → Linked Devices → Link a Device → Scan the code.
- **Done!** Send `Hi` to the bot (or yourself) on WhatsApp.

> On first run, a **QR code** will appear in your terminal. Open WhatsApp → Linked Devices → Scan QR code.

---

## 🔑 Getting Your Free Gemini API Key

1. Visit [aistudio.google.com/app/apikey](https://aistudio.google.com/app/apikey)
2. Sign in with your Google account
3. Click **"Create API Key"**
4. Copy the key and set it: `export GEMINI_API_KEY="your_key"`

**No credit card required for the free tier.**

---

## 📊 Free Tier Model Limits (Feb 2026)

StudyClaw uses **`gemini-2.0-flash-lite`** by default — the most generous free model.

| Model | RPM | RPD | Used For |
|:------|:---:|:---:|:---------|
| `gemini-2.0-flash-lite` ✅ | 15 | **1,000** | Default chat & replies |
| `gemini-2.5-flash` ✅ | 10 | 250 | Quiz generation |
| `gemini-3-pro` / `gemini-3.1-pro` ❌ | — | — | **Not available via free API key** |
| Thinking models ❌ | — | — | **Not available via free API key** |

A built-in **token-bucket rate limiter** (12 RPM headroom) prevents 429 errors, with automatic retry.

---

## 💬 Example Usage

After linking WhatsApp, send these messages to the bot:

| Message | Response |
|:--------|:---------|
| `Hi` | Introduces itself and its capabilities |
| `Summarize my thermodynamics notes` | Retrieves and summarizes relevant notes from the DB |
| `Quiz me on Circuit Theory` | Sends 3–5 MCQ questions |
| `Show a diagram of the OSI model` | Replies with a Mermaid diagram + viewer link |
| (Send a PDF) | Bot indexes the PDF into the local knowledge base |

---

## ⚙️ Configuration

The optional `config.json` is auto-generated by the setup wizard (`go run cmd/setup.go`), or you can create it manually:

```json
{
  "gemini_api_key": "your_key_here",
  "owner_jid": "919876543210@s.whatsapp.net",
  "model_name": "gemini-2.0-flash-lite",
  "db_path": "studyclaw.db",
  "session_path": "whatsapp_session.db"
}
```

Alternatively, you can skip the setup wizard by creating a `.env` file in the root directory or exporting these environment variables:
```bash
export GEMINI_API_KEY="your_key"
export TELEGRAM_BOT_TOKEN="your_tg_token"
export LLM_PROVIDER="gemini" # Options: gemini, codex (ChatGPT Pro), antigravity
export STUDYCLAW_OWNER_NUMBER="91XXXXXXXXXX"
```

---

## 📦 Resource Usage

| Component | RAM |
|:----------|:----|
| Go binary | ~15–30 MB |
| SQLite database | ~5–20 MB |
| **Total** | **~30–50 MB** |

Runs comfortably on any phone with **4 GB+ RAM**.

---

## 🗺️ Roadmap

- [x] Phase 1: WhatsApp bridge, PDF indexing, daily quizzes, agent loop
- [x] Phase 2: Gemini native tool calling, diagram viewer, Google Drive sync
- [x] Phase 3: Rate-limit protection, free-tier model optimization, clean architecture
- [x] Phase 4: Self-Reflection memory, Academic Calendar Engine, Report Tools
- [ ] Phase 5: GTU PYQ predictor, Telegram Mini App, Handwriting OCR
- [ ] Phase 6: Exam countdown alerts, multi-user group support

---

## 📄 License

MIT — fork freely, learn boldly 🦞
