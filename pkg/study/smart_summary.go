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

// groupNoteThreshold defines the minimum length for auto-saving from a passive group.
const groupNoteThreshold = 200

// Process decides how to handle a raw message.
// Returns: (reply, shouldContinueToAgent)
// isGroupMessage: true when from a passive-mode group channel (not the user directly)
func (h *SmartMessageHandler) Process(ctx context.Context, content string) (string, bool) {
	return h.ProcessWithSource(ctx, content, false)
}

// ProcessWithSource separates group monitoring logic from direct bot interaction.
func (h *SmartMessageHandler) ProcessWithSource(ctx context.Context, content string, isGroupMessage bool) (string, bool) {
	trimmed := strings.TrimSpace(content)

	// Empty or command messages — always pass through
	if trimmed == "" || strings.HasPrefix(trimmed, "!") || strings.HasPrefix(trimmed, "/") {
		return "", true
	}

	// Direct user→bot messages: ALWAYS forward to the agent, never block.
	if !isGroupMessage {
		return "", true
	}

	// Group passive monitoring: short messages are silently noted
	if len(trimmed) <= groupNoteThreshold {
		_ = h.db.SaveNote("group_note", trimmed, "group")
		return "", false // silent save, no reply to group
	}

	// Group passive monitoring: long messages are AI-summarized
	summary, err := h.summarize(ctx, trimmed)
	if err != nil {
		return "", false // silently skip on error
	}

	reply := fmt.Sprintf("📋 *Key Points:*\n%s\n\nReply *more* for full details.", summary)
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
