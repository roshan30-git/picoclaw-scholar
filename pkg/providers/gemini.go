package providers

import (
	"context"
	"fmt"
	"strings"

	"github.com/user/studyclaw/pkg/tools"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	client *genai.Client
	model  string
}

func NewGeminiProvider(apiKey string) (*GeminiProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("gemini API key is required")
	}
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}
	return &GeminiProvider{client: client, model: "gemini-2.0-flash"}, nil
}

func (g *GeminiProvider) GetDefaultModel() string {
	return g.model
}

func (g *GeminiProvider) Chat(
	ctx context.Context,
	messages []tools.Message,
	toolDefs []tools.ToolDefinition,
	model string,
	options map[string]any,
) (*tools.LLMResponse, error) {
	if model == "" {
		model = g.model
	}

	m := g.client.GenerativeModel(model)

	var parts []genai.Part
	for _, msg := range messages {
		parts = append(parts, genai.Text(msg.Role+": "+msg.Content))
	}

	resp, err := m.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, fmt.Errorf("gemini generate: %w", err)
	}

	var sb strings.Builder
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				sb.WriteString(fmt.Sprintf("%v", part))
			}
		}
	}

	return &tools.LLMResponse{
		Content: sb.String(),
	}, nil
}
