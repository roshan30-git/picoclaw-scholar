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

// ModelRouter dynamically selects the most cost-efficient model for a given task.
// Implements tools.LLMProvider.
type ModelRouter struct {
	baseProvider tools.LLMProvider
}

// NewModelRouter creates a dynamic LLM provider wrapper
func NewModelRouter(provider tools.LLMProvider) *ModelRouter {
	return &ModelRouter{
		baseProvider: provider,
	}
}

// Chat intercepts the standard Chat call and routes it to the correct model tier.
func (m *ModelRouter) Chat(ctx context.Context, history []tools.Message, toolsDef []tools.ToolDefinition, requestedModel string, options map[string]any) (*tools.LLMResponse, error) {
	targetModel := string(TierStandard) // Default to flash

	if requestedModel != "" {
		targetModel = requestedModel
	}

	return m.baseProvider.Chat(ctx, history, toolsDef, targetModel, options)
}

// GetDefaultModel satisfies tools.LLMProvider.
func (m *ModelRouter) GetDefaultModel() string {
	return string(TierStandard)
}

// RouteCost returns the selected tier based on a hint.
func (m *ModelRouter) RouteCost(taskHint string) string {
	switch taskHint {
	case "indexer", "scheduler":
		return string(TierFree)
	case "strategist", "deep_analysis":
		return string(TierAdvanced)
	default:
		return string(TierStandard)
	}
}
