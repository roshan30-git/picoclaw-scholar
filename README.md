<div align="center">
  <h1>🦞 PicoClaw: Scholar Edition</h1>
  <h3>A lightweight, WhatsApp/Telegram-native AI study agent that passively indexes your notes and proactively drills you based on real academic context.</h3>

  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white">
  <img src="https://img.shields.io/badge/RAM-50MB--target-brightgreen">
  <img src="https://img.shields.io/badge/Platform-Android%20%7C%20Termux-orange">
  <img src="https://img.shields.io/badge/license-MIT-green">
  <img src="https://img.shields.io/badge/Powered_by-Gemini_Flash-blue?logo=google">
</div>

---

> **StudyClaw** is a fork of [PicoClaw](https://github.com/sipeed/picoclaw) — purpose-built for students. It runs on your $30 Android phone via Termux and turns your WhatsApp/Telegram study group into an autonomous AI mentor.

## ✨ Features (MVP — Phase 1)

| Feature | Description |
|---------|-------------|
| 📥 **Passive Note Indexing** | Listens to your WhatsApp study group. Auto-detects PDFs & images. Uses Gemini OCR to store notes locally in SQLite. |
| 🎯 **Proactive Daily Quizzes** | Sends 3–5 MCQs at 8 PM every day, weighted by your university PYQs. |
| 📅 **Calendar-Aware Mode** | Switches to Exam Mode 14 days before finals. Relaxes to Festival Mode on holidays. |
| 📊 **Telegram Diagram Viewer** | Opens interactive diagrams (circuits, flowcharts, graphs) via Telegram Mini App. |
| ☁️ **Google Drive Library** | Syncs your textbooks and lecture slides from a shared Drive folder auto-matically. |
| 🤖 **Single-Call Intelligence** | All personas (Peer, Drill Sergeant, Librarian) are handled in **one Gemini call** via dynamic prompt overlays — no multi-agent overhead. |

## 🏗️ Architecture

```
PicoClaw Gateway (Go, <15MB RAM)
├── channels/
│   ├── whatsapp/   ← whatsmeow (message events, group indexing)
│   └── telegram/   ← go-telegram-bot-api (Mini App diagrams)
├── integrations/
│   └── gdrive/     ← Google Drive (books, lecture slides)
├── media/
│   ├── pdf/        ← pdfcpu (text extraction)
│   ├── ocr/        ← Tesseract via Termux exec
│   └── diagram/    ← Mermaid → Telegram Mini App
├── ai/
│   └── gemini.go   ← Gemini 1.5 Flash (BYOK, <300K tokens/day)
├── database/
│   ├── sqlite.go
│   └── vector.go   ← sqlite-vec for local RAG
├── scheduler/
│   └── cron.go     ← robfig/cron (daily quiz, exam countdown)
└── workspace/
    ├── PROMPTS/    ← base_soul.md + mode overlays
    └── MEMORY/     ← syllabus, calendar, notes
```

## 🚀 Quick Start (Termux on Android)

```bash
# 1. Install Termux from F-Droid, then:
pkg install golang git tesseract nodejs

# 2. Clone this repo
git clone https://github.com/YOUR_USERNAME/picoclaw-scholar.git
cd picoclaw-scholar

# 3. Build
go build -o studyclaw ./cmd/main.go

# 4. Configure
cp config.json ~/.studyclaw/config.json
nano ~/.studyclaw/config.json  # Paste your Gemini API key

# 5. Run (Scan WhatsApp QR on first launch)
./studyclaw
```

## ⚙️ Configuration

Edit `~/.studyclaw/config.json`:

```json
{
  "gemini": { "api_key": "YOUR_KEY_HERE" },
  "whatsapp": { "owner_number": "+91XXXXXXXXXX" },
  "scheduler": { "daily_quiz_time": "20:00" }
}
```

Get a free Gemini API key at [aistudio.google.com](https://aistudio.google.com/app/apikey) (1M tokens/day free).

## 📊 Resource Usage

| Component | RAM |
|-----------|-----|
| Go binary (StudyClaw) | ~15–30 MB |
| SQLite + sqlite-vec | ~5–20 MB |
| Tesseract (on-demand) | ~50 MB peak |
| **Total resident** | **~50 MB** |

Runs comfortably on **4 GB RAM** phones. Recommended: **6–8 GB** for smooth concurrent indexing.

## 🗺️ Roadmap

- [x] Phase 1: WhatsApp bridge, note indexing, daily quizzes
- [ ] Phase 2: GTU PYQ scraper, Telegram Mini App diagrams, Google Drive sync
- [ ] Phase 3: Handwriting OCR, multi-persona prompt routing, launch

## 📄 License

MIT — fork freely, learn boldly 🦞
