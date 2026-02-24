package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/roshan30-git/picoclaw-scholar/integrations/whatsapp"
	"github.com/roshan30-git/picoclaw-scholar/integrations/gdrive"
	pkgdb "github.com/roshan30-git/picoclaw-scholar/pkg/database"
	"github.com/roshan30-git/picoclaw-scholar/pkg/agent"
	"github.com/roshan30-git/picoclaw-scholar/pkg/bus"
	"github.com/roshan30-git/picoclaw-scholar/pkg/channels"
	"github.com/roshan30-git/picoclaw-scholar/pkg/config"
	"github.com/roshan30-git/picoclaw-scholar/pkg/providers"
	"github.com/roshan30-git/picoclaw-scholar/pkg/study"
	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
)

func main() {
	fmt.Println("🦞 PicoClaw: Scholar Edition — Initializing...")

	// 1. Load config (minimal impl for now)
	ctx := context.Background()
	
	// 2. Initialize Database
	db, err := pkgdb.New("studyclaw.db") // Changed to local relative path for testing
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	// Note: db.Close() would go here but we just use the reference.

	// 3. Initialize Google Drive (for books/syllabus)
	driveClient, err := gdrive.New(ctx)
	if err != nil {
		log.Printf("Warning: Google Drive not linked: %v", err)
	} else {
		fmt.Println("✅ Google Drive linked.")
		_ = driveClient // Use for background indexing later
	}

	// 4. Initialize Message Bus & Agent Loop
	msgBus := bus.NewMessageBus()
	provider, err := providers.NewGeminiProvider(os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		log.Printf("Warning: Gemini provider init failed (API key missing?): %v", err)
	}
	
	// Create channels manager
	chMgr := channels.NewManager()

	// Initialize Agent loop
	if provider != nil {
		agentLoop := agent.NewAgentLoop(config.DefaultConfig(), msgBus, provider)
		agentLoop.RegisterTool(study.NewQuizTool(study.NewQuizEngine(provider, db)))
		agentLoop.RegisterTool(study.NewIngestTool(study.NewIngestionEngine(db)))
		agentLoop.SetChannelManager(chMgr)
		go agentLoop.Run(ctx)
		fmt.Println("🤖 Agent Loop initialized.")
	}

	// 5. Initialize WhatsApp Bridge
	waClient, err := whatsapp.New("whatsapp_session.db", msgBus)
	if err != nil {
		log.Fatalf("Failed to init WhatsApp: %v", err)
	}
	chMgr.Register(waClient)

	// Start all channels
	if err := chMgr.StartAll(ctx); err != nil {
		log.Fatalf("Failed to start channels: %v", err)
	}
	defer chMgr.StopAll(ctx)

	fmt.Println("🚀 StudyClaw is alive! Ready for messages...")

	// Wait for interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n👋 StudyClaw is shutting down...")
}
