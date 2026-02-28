package providers

import (
	"context"

	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
)

type ModelTier string

const (
	TierFree     ModelTier = "gemini-2.0-flash-lite" // Free, fast, ideal for basic routing
	TierStandard ModelTier = "gemini-2.0-flash"      // Fast, multimodal, cheap, ideal for visual/OCR/base agent
	TierAdvanced ModelTier = "gemini-2.5-pro"        // Slower, expensive, high reasoning, ideal for PYQ analysis
)

// ModelRouter dynamically selects the most cost-efficient model for a given task
type ModelRouter struct {
	baseProvider tools.LLMProvider
}

// NewModelRouter creates a dynamic LLM provider wrapper
func NewModelRouter(provider tools.LLMProvider) *ModelRouter {
	return &ModelRouter{
		baseProvider: provider,
	}
}

// Chat intercepts the standard Chat call and routes it to the correct model tier based on the persona/task.
// In this scaffold, we override the requested model if we know a cheaper one suffices.
func (m *ModelRouter) Chat(ctx context.Context, history []tools.Message, toolsDef []tools.ToolDefinition, requestedModel string, config *tools.ChatConfig) (*tools.ChatResponse, error) {
	targetModel := string(TierStandard) // Default to flash

	// Heuristics for overriding requested models to save cost:
	// 1. If it's a simple scheduler task, flash-lite might suffice (but stick to flash for tool reliability MVP)
	// 2. If it's the "Exam Strategist" or dealing with large PDFs, use Pro natively.
	// For MVP, we let the explicitly requested model pass unless we build strict overrides.
	
	if requestedModel != "" {
		targetModel = requestedModel
	}

	return m.baseProvider.Chat(ctx, history, toolsDef, targetModel, config)
}

// RouteCost returns the selected tier based on a hint.
func (m *ModelRouter) RouteCost(taskHint string) string {
	switch taskHint {
	case "indexer", "scheduler":
		return string(TierFree) // Could use flash-lite if reliable enough for tools
	case "strategist", "deep_analysis":
		return string(TierAdvanced)
	default:
		return string(TierStandard)
	}
}
