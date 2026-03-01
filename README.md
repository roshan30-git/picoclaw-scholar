<div align="center">
  <h1>🦞 StudyClaw</h1>
  <h3>Your autonomous AI study companion — lives in WhatsApp & Telegram</h3>
  <br/>
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white">
  <img src="https://img.shields.io/badge/Platform-Windows%20%7C%20Android%20(Termux)-orange">
  <img src="https://img.shields.io/badge/AI-Gemini%202.0%20Flash-blue?logo=google">
  <img src="https://img.shields.io/badge/DB-SQLite-003B57?logo=sqlite">
</div>

---

**StudyClaw** is a Go-based agentic AI bot that runs directly on your Windows PC or Android phone (via Termux). It connects to WhatsApp and Telegram, indexes your study materials, and proactively sends you quizzes, summaries, and diagrams — powered by **Gemini**.

---

## ✨ Key Features

| Feature | Description |
|:--------|:------------|
| 📥 **Instant Indexing** | Send a PDF or text to the bot. It's instantly indexed for retrieval. |
| 🎯 **Adaptive Quizzes** | Get 3–5 MCQs based on your actual notes, tailored to your learning pace. |
| 📊 **Proactive Diagrams** | AI generates Mermaid diagrams; view them at `http://localhost:8080`. |
| 📅 **Academic Soul** | The bot's personality shifts (Drill Sergeant vs. Peer) based on your exam dates. |
| 🦾 **Native Tool Calling** | Uses Gemini's function calling to autonomously search notes or add deadlines. |

---

## 🚀 Quick Start (Windows) — Recommended

StudyClaw is optimized for Windows with a **one-click launcher**.

1.  **Install Go**: Download from [golang.org/dl](https://golang.org/dl/).
2.  **Clone & Run**:
    ```powershell
    git clone https://github.com/roshan30-git/picoclaw-scholar.git
    cd picoclaw-scholar
    .\run.ps1
    ```
3.  **Setup**: Follow the on-screen wizard to enter your Gemini API key.
4.  **Scan**: A WhatsApp QR code will appear. Scan it with your phone. **Done!**

---

## 🚀 Quick Start (Termux on Android)

1.  **Install Environment**:
    ```bash
    pkg update && pkg upgrade -y
    pkg install golang git clang make -y
    ```
2.  **Setup**:
    ```bash
    git clone https://github.com/roshan30-git/picoclaw-scholar.git
    cd picoclaw-scholar
    go mod tidy
    ```
3.  **Run**:
    ```bash
    export GEMINI_API_KEY="your_key"
    go run cmd/main.go
    ```

---

## 🔑 Getting Your API Key
Get your free Gemini API key at [aistudio.google.com](https://aistudio.google.com/apikey). No credit card required.

---

## 🛠️ Configuration
The app uses a `.env` file for secrets. The Windows launcher (`run.ps1`) creates this for you automatically.

```env
GEMINI_API_KEY=AIza...
TELEGRAM_BOT_TOKEN=...
STUDYCLAW_OWNER_NUMBER=91...
LLM_PROVIDER=gemini
```

---

## 🗺️ Roadmap
- [x] WhatsApp & Telegram Integration
- [x] Gemini Tool Calling & PDF Ingestion
- [x] Windows One-Click Launcher
- [x] Academic Calendar & Reflection Engine
- [ ] Handwriting OCR support
- [ ] Exam Countdown Alerts

---

## 📄 License
MIT — Learn boldly 🦞
