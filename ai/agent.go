package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/genai"
	"github.com/user/studyclaw/database"
)

type Agent struct {
	db     *database.DB
	apiKey string
}

func NewAgent(db *database.DB, apiKey string) (*Agent, error) {
	return &Agent{db: db, apiKey: apiKey}, nil
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

	// 3. Call Gemini Flash API
	if a.apiKey == "" {
		return "⚠️ API Key not found. Please add 'gemini.api_key' to ~/.studyclaw/config.json", nil
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: a.apiKey})
	if err != nil {
		return "", fmt.Errorf("genai client: %w", err)
	}

	// Combine system prompt + user message
	fullPrompt := prompt + "\n\nUser: " + text
	// If mediaPath is present, we would add it here (TODO)

	resp, err := client.Models.GenerateContent(ctx, "gemini-1.5-flash", genai.Text(fullPrompt), nil)
	if err != nil {
		return "", fmt.Errorf("generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "⚠️ No response from Gemini.", nil
	}

	// Extract text response
	// Assuming Part is of type Text
	part := resp.Candidates[0].Content.Parts[0]
	return part.Text, nil
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
		// Check if we are running from a subdirectory (e.g. tests)
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			basePath = "../workspace/PROMPTS/base_soul.md"
		}
	}

	base, err := os.ReadFile(basePath)
	if err != nil {
		return "", fmt.Errorf("load base prompt: %w", err)
	}

	return string(base) + "\n\n## Current Mode: " + mode, nil
}
