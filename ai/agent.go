package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/studyclaw/database"
)

type Agent struct {
	db *database.DB
}

func NewAgent(db *database.DB) (*Agent, error) {
	return &Agent{db: db}, nil
}

// Process handles an incoming message and returns a personalized AI response.
func (a *Agent) Process(ctx context.Context, sender, text, mediaPath string) (string, error) {
	// 1. Determine current mode (Tutor, Quiz, Alert) based on calendar/history
	mode := a.determineMode()

	// 2. Load the base soul + mode overlay
	prompt, err := a.buildPrompt(mode)
	if err != nil {
		return "", err
	}

	// 3. (Mock) Call Gemini Flash API
	// In a real implementation, this would use google.golang.org/genai
	fmt.Printf("[AI System Prompt Log]: Using %s mode\n", mode)
	
	// Simple mock logic for now
	if strings.Contains(strings.ToLower(text), "quiz") {
		return "📝 Ready for your daily drill? Here's a question based on your 'Circuit Theory' notes from yesterday: What is the primary difference between a mesh and a loop?", nil
	}

	return "🦞 StudyClaw here! I've indexed your latest notes. Feeling ready for next week's Mock Exam?", nil
}

func (a *Agent) determineMode() string {
	// Logic to check calendar for exams/festivals
	// For now, default to 'tutor'
	return "tutor"
}

func (a *Agent) buildPrompt(mode string) (string, error) {
	home, _ := os.UserHomeDir()
	basePath := filepath.Join(home, ".studyclaw/workspace/PROMPTS/base_soul.md")
	
	// Fallback to local project path if not in home
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		basePath = "workspace/PROMPTS/base_soul.md"
	}

	base, err := os.ReadFile(basePath)
	if err != nil {
		return "", fmt.Errorf("load base prompt: %w", err)
	}

	return string(base) + "\n\n## Current Mode: " + mode, nil
}
