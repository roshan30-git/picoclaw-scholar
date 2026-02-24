package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/studyclaw/integrations/whatsapp"
	"github.com/user/studyclaw/integrations/gdrive"
	"github.com/user/studyclaw/ai"
	"github.com/user/studyclaw/config"
	"github.com/user/studyclaw/database"
)

func main() {
	fmt.Println("🦞 PicoClaw: Scholar Edition — Initializing...")

	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()
	
	// 2. Initialize Database
	db, err := database.New("~/.studyclaw/studyclaw.db")
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	// 3. Initialize AI Agent
	agent, err := ai.NewAgent(db, cfg.Gemini.APIKey)
	if err != nil {
		log.Fatalf("Failed to init AI agent: %v", err)
	}

	// 4. Initialize Google Drive (for books/syllabus)
	driveClient, err := gdrive.New(ctx)
	if err != nil {
		log.Printf("Warning: Google Drive not linked: %v", err)
	} else {
		fmt.Println("✅ Google Drive linked.")
		_ = driveClient // Use for background indexing later
	}

	// 5. Initialize WhatsApp Bridge
	handler := func(sender, text, mediaPath string) {
		fmt.Printf("Message from %s: %s\n", sender, text)
		response, err := agent.Process(ctx, sender, text, mediaPath)
		if err != nil {
			log.Printf("AI error: %v", err)
			return
		}
		// Send response back via WhatsApp (this would be wired to waClient.Send)
		fmt.Printf("AI Response: %s\n", response)
	}

	waClient, err := whatsapp.New("~/.studyclaw/whatsapp_session.json", handler)
	if err != nil {
		log.Fatalf("Failed to init WhatsApp: %v", err)
	}
	defer waClient.Disconnect()

	fmt.Println("🚀 StudyClaw is alive! Scanning for changes...")

	// Wait for interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n👋 StudyClaw is shutting down...")
}
