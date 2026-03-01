package study

import (
	"context"
	"fmt"
	"strings"

	"github.com/roshan30-git/picoclaw-scholar/pkg/database"
	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
)

const shortMessageThreshold = 120

// SmartMessageHandler pre-processes raw messages before the main agent pipeline.
// Short messages are silently saved. Long messages are AI-summarized.
type SmartMessageHandler struct {
	provider tools.LLMProvider
	db       *database.DB
}

// NewSmartMessageHandler creates a handler backed by an LLM and a database.
func NewSmartMessageHandler(provider tools.LLMProvider, db *database.DB) *SmartMessageHandler {
	return &SmartMessageHandler{provider: provider, db: db}
}

// Process decides how to handle a raw message.
// Returns: (reply, shouldContinueToAgent)
func (h *SmartMessageHandler) Process(ctx context.Context, content string) (string, bool) {
	trimmed := strings.TrimSpace(content)

	// Empty or command messages — pass through as-is
	if trimmed == "" || strings.HasPrefix(trimmed, "!") || strings.HasPrefix(trimmed, "/") {
		return "", true
	}

	// Short message: save silently and acknowledge
	if len(trimmed) <= shortMessageThreshold {
		_ = h.db.SaveNote(trimmed, trimmed, "whatsapp")
		return "📌 Saved!", false
	}

	// Long message: summarize into 3 key points
	summary, err := h.summarize(ctx, trimmed)
	if err != nil {
		return "", true // fallback to full agent on error
	}

	reply := fmt.Sprintf("📋 *Key Points:*\n%s\n\nWant full details? Reply with *more*", summary)
	return reply, false
}

func (h *SmartMessageHandler) summarize(ctx context.Context, content string) (string, error) {
	prompt := fmt.Sprintf(
		"Extract exactly 3 concise bullet points from this message. "+
			"Use '• ' as the bullet symbol. Be brief:\n\n%s",
		content,
	)
	resp, err := h.provider.Chat(ctx, []tools.Message{{Role: "user", Content: prompt}}, nil, "gemini-2.0-flash", nil)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}
