package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/roshan30-git/picoclaw-scholar/integrations/gdrive"
	"github.com/roshan30-git/picoclaw-scholar/integrations/whatsapp"
	"github.com/roshan30-git/picoclaw-scholar/pkg/agent"
	"github.com/roshan30-git/picoclaw-scholar/pkg/bus"
	"github.com/roshan30-git/picoclaw-scholar/pkg/channels"
	"github.com/roshan30-git/picoclaw-scholar/pkg/config"
	pkgdb "github.com/roshan30-git/picoclaw-scholar/pkg/database"
	"github.com/roshan30-git/picoclaw-scholar/pkg/providers"
	"github.com/roshan30-git/picoclaw-scholar/pkg/study"
	"github.com/roshan30-git/picoclaw-scholar/pkg/viewer"
)

func main() {
	fmt.Println("🦞 StudyClaw — Initializing...")

	ctx := context.Background()

	// 1. Initialize Database
	db, err := pkgdb.New("studyclaw.db")
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// 2. Initialize Google Drive (optional)
	driveClient, err := gdrive.New(ctx)
	if err != nil {
		log.Printf("Warning: Google Drive not linked: %v", err)
	} else {
		fmt.Println("✅ Google Drive linked.")
		_ = driveClient
	}

	// 3. Initialize Message Bus
	msgBus := bus.NewMessageBus()

	// 4. Initialize Gemini Provider (free-tier optimized: gemini-2.0-flash-lite)
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	provider, err := providers.NewGeminiProvider(geminiAPIKey)
	if err != nil {
		log.Printf("⚠️  Gemini provider not available (set GEMINI_API_KEY env var): %v", err)
	}

	// 5. Diagram Viewer Server (localhost:8080)
	viewerSrv := viewer.NewServer(8080)
	go viewerSrv.Start()
	fmt.Println("📊 Diagram viewer available at http://127.0.0.1:8080")

	// 6. Channel Manager
	chMgr := channels.NewManager()

	// 7. Agent Loop (only started if Gemini is available)
	if provider != nil {
		agentLoop := agent.NewAgentLoop(config.DefaultConfig(), msgBus, provider)
		agentLoop.RegisterTool(study.NewQuizTool(study.NewQuizEngine(provider, db)))
		agentLoop.RegisterTool(study.NewIngestTool(study.NewIngestionEngine(db)))
		agentLoop.SetChannelManager(chMgr)
		go agentLoop.Run(ctx)
		fmt.Println("🤖 Agent Loop initialized with gemini-2.0-flash-lite (free tier)")
	}

	// 8. WhatsApp Channel
	waClient, err := whatsapp.New("whatsapp_session.db", msgBus)
	if err != nil {
		log.Fatalf("Failed to init WhatsApp: %v", err)
	}
	chMgr.Register(waClient)

	// 9. Start all channels
	if err := chMgr.StartAll(ctx); err != nil {
		log.Fatalf("Failed to start channels: %v", err)
	}
	defer chMgr.StopAll(ctx)

	fmt.Println("🚀 StudyClaw is alive! Send a message via WhatsApp to start.")

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n👋 StudyClaw is shutting down...")
}
