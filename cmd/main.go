package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/roshan30-git/picoclaw-scholar/integrations/gdrive"
	"github.com/roshan30-git/picoclaw-scholar/integrations/whatsapp"
	"github.com/roshan30-git/picoclaw-scholar/pkg/agent"
	"github.com/roshan30-git/picoclaw-scholar/pkg/bus"
	"github.com/roshan30-git/picoclaw-scholar/pkg/channels"
	"github.com/roshan30-git/picoclaw-scholar/pkg/channels/telegram"
	"github.com/roshan30-git/picoclaw-scholar/pkg/config"
	pkgdb "github.com/roshan30-git/picoclaw-scholar/pkg/database"
	"github.com/roshan30-git/picoclaw-scholar/pkg/memory"
	"github.com/roshan30-git/picoclaw-scholar/pkg/providers"
	"github.com/roshan30-git/picoclaw-scholar/pkg/setup"
	"github.com/roshan30-git/picoclaw-scholar/pkg/study"
	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
	"github.com/roshan30-git/picoclaw-scholar/pkg/viewer"
	"github.com/roshan30-git/picoclaw-scholar/pkg/visual"
)

func main() {
	_ = godotenv.Load()
	fmt.Println("🦞 StudyClaw — Initializing...")

	// Launch local web UI if critical configuration is missing
	setup.RunServerIfConfigMissing()

	cfg := config.LoadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize Database
	db, err := pkgdb.New("studyclaw.db")
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// 2. Initialize Google Drive (strictly silent & optional)
	var driveClient *gdrive.Client
	enableDrive := os.Getenv("STUDYCLAW_ENABLE_GDRIVE") == "true"
	if enableDrive {
		home, _ := os.UserHomeDir()
		credPath := filepath.Join(home, ".studyclaw/google_credentials.json")
		if _, err := os.Stat(credPath); err == nil {
			driveClient, err = gdrive.New(ctx)
			if err != nil {
				log.Printf("Warning: Failed to init Google Drive: %v", err)
			} else {
				fmt.Println("✅ Google Drive linked.")
			}
		}
	}
	_ = driveClient

	// 3. Initialize Message Bus
	msgBus := bus.NewMessageBus()

	// 4. Initialize LLM Provider (antigravity, codex, or gemini)
	providerType := os.Getenv("LLM_PROVIDER")
	var provider tools.LLMProvider

	switch providerType {
	case "antigravity":
		fmt.Println("🚀 Using Antigravity Provider (Cloud Code Assist)")
		provider = providers.NewAntigravityProvider()
	case "codex":
		fmt.Println("🚀 Using Codex Provider (ChatGPT Pro)")
		openaiKey := os.Getenv("OPENAI_API_KEY")
		accountID := os.Getenv("CHATGPT_ACCOUNT_ID")
		provider = providers.NewCodexProvider(openaiKey, accountID)
	default:
		geminiAPIKey := os.Getenv("GEMINI_API_KEY")
		provider, err = providers.NewGeminiProvider(geminiAPIKey)
		if err != nil {
			log.Printf("⚠️  Gemini provider not available (set GEMINI_API_KEY env var): %v", err)
		}
	}

	// 5. Diagram Viewer Server (localhost:8080)
	var allowedOrigins []string
	if cfg.TelegramWebAppURL == "" {
		log.Printf("⚠️  TelegramWebAppURL not set; viewer CORS will be restricted (no origin allowed)")
	} else if u, err := url.Parse(cfg.TelegramWebAppURL); err != nil || u.Scheme == "" || u.Host == "" {
		log.Printf("⚠️  Failed to parse TelegramWebAppURL %q: viewer CORS will be restricted", cfg.TelegramWebAppURL)
	} else {
		allowedOrigins = append(allowedOrigins, fmt.Sprintf("%s://%s", u.Scheme, u.Host))
	}

	viewerSrv := viewer.NewServer(8080, allowedOrigins)
	go viewerSrv.Start()
	fmt.Println("📊 Diagram viewer available at http://127.0.0.1:8080")
	visManager := visual.NewManager(viewerSrv)

	// 7. Channel Manager
	chMgr := channels.NewManager()

	// Initialize Phase 4 Engines
	calendarEngine := study.NewCalendarEngine()
	reflectionManager := memory.NewReflectionManager("workspace")
	personaRouter := agent.NewPersonaRouter()
	deadlineTracker := study.NewDeadlineTracker(db)
	weeklyCards := study.NewWeeklyCardsGenerator(db, provider, msgBus, os.Getenv("STUDYCLAW_OWNER_NUMBER"))

	// 6b. Start proactive cron scheduler
	ownerID := os.Getenv("STUDYCLAW_OWNER_NUMBER")
	scheduler := study.NewScheduler(deadlineTracker, weeklyCards, msgBus, ownerID)
	scheduler.ScheduleReminders()
	scheduler.ScheduleWeeklyCards()
	go scheduler.Start(ctx)
	fmt.Println("📅 Proactive scheduler started (daily reminders + weekly flashcards)")

	// 8. Agent Loop (only started if LLM is available)
	if provider != nil {
		agentLoop := agent.NewAgentLoop(cfg, msgBus, provider, visManager, personaRouter, calendarEngine, reflectionManager, db)
		agentLoop.RegisterTool(study.NewQuizTool(study.NewQuizEngine(provider, db)))
		agentLoop.RegisterTool(study.NewIngestTool(study.NewIngestionEngine(db)))
		agentLoop.RegisterTool(study.NewSearchNotesTool(db))
		agentLoop.RegisterTool(study.NewAddDeadlineTool(deadlineTracker))
		agentLoop.RegisterTool(study.NewViewDeadlinesTool(deadlineTracker))
		agentLoop.RegisterTool(tools.NewReportGeneratorTool(provider))
		agentLoop.RegisterTool(tools.NewWebSearchTool())
		agentLoop.RegisterTool(tools.NewExcelTool())
		agentLoop.RegisterTool(tools.NewDiagramTool())
		agentLoop.SetChannelManager(chMgr)
		agentLoop.SetOnShutdown(cancel)
		go agentLoop.Run(ctx)
		fmt.Println("🤖 Agent Loop initialized with current LLM provider")
	}

	// 9. Channels (WhatsApp & Telegram)
	// Initialize OCR Pipeline for WhatsApp (using current provider)
	ocrPipeline, err := study.NewOCRPipeline(os.Getenv("GEMINI_API_KEY"), db)
	if err != nil {
		log.Printf("Warning: Failed to init OCR Pipeline: %v", err)
	}

	waClient, err := whatsapp.New(ctx, "whatsapp_session.db", msgBus, cfg.AllowedGroupJIDs, cfg.PassiveGroupJIDs, ocrPipeline)
	if err != nil {
		log.Printf("Warning: Failed to init WhatsApp: %v", err)
	} else {
		chMgr.Register(waClient)
	}

	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if telegramToken != "" {
		cfg.Channels.Telegram.Token = telegramToken
		cfg.Channels.Telegram.Enabled = true

		tgClient, err := telegram.NewTelegramChannel(cfg, msgBus, db)
		if err != nil {
			log.Printf("Warning: Failed to init Telegram channel: %v", err)
		} else {
			chMgr.Register(tgClient)
			fmt.Println("✅ Telegram bot linked.")
		}
	}

	// 10. Start all channels
	if err := chMgr.StartAll(ctx); err != nil {
		log.Fatalf("Failed to start channels: %v", err)
	}
	defer chMgr.StopAll(ctx)

	fmt.Println("🚀 StudyClaw is alive! Send a message via WhatsApp to start.")
	fmt.Println("   (Type 'stop' or 'exit' in terminal to shut down)")

	// 11. Terminal Command Listener
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if text == "stop" || text == "exit" {
				fmt.Println("\n🛑 Shutdown command received from terminal...")
				cancel()
				return
			}
		}
	}()

	// Graceful shutdown logic remains for signals and context cancellation
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n👋 StudyClaw is shutting down...")
}
