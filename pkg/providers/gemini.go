package providers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
	"google.golang.org/genai"
)

// Free tier model selection:
// - gemini-2.0-flash-lite: 15 RPM, 1000 RPD, 250K TPM — best value for a personal bot
// - gemini-2.0-flash:       10 RPM,  200 RPD, 250K TPM — fallback for reasoning tasks
// - gemini-2.5-flash:       10 RPM,  250 RPD, 250K TPM — latest stable, better reasoning
const (
	ModelDefault   = "gemini-2.0-flash-lite" // always use this for chat (most generous limits)
	ModelReasoning = "gemini-2.0-flash"       // use for quiz/complex tasks
)

type GeminiProvider struct {
	client *genai.Client
	model  string
}

func NewGeminiProvider(apiKey string) (*GeminiProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}
	return &GeminiProvider{client: client, model: ModelDefault}, nil
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

	// Build the config
	cfg := &genai.GenerateContentConfig{}

	// Attach tools if provided
	if len(toolDefs) > 0 {
		var decls []*genai.FunctionDeclaration
		for _, td := range toolDefs {
			decls = append(decls, &genai.FunctionDeclaration{
				Name:        td.Function.Name,
				Description: td.Function.Description,
				Parameters:  parseSchema(td.Function.Parameters),
			})
		}
		cfg.Tools = []*genai.Tool{{FunctionDeclarations: decls}}
	}

	// Build the message history as genai.Content slice
	var contents []*genai.Content
	for _, msg := range messages {
		role := msg.Role
		if role == "model" {
			role = "model"
		} else {
			role = "user"
		}
		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: []*genai.Part{genai.NewPartFromText(msg.Content)},
		})
	}

	// Retry once on rate-limit errors (429)
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(5 * time.Second)
		}

		resp, err := g.client.Models.GenerateContent(ctx, model, contents, cfg)
		if err != nil {
			lastErr = err
			if strings.Contains(err.Error(), "429") {
				continue
			}
			return nil, fmt.Errorf("gemini generate: %w", err)
		}

		var contentBuilder strings.Builder
		var toolCalls []tools.ToolCall
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.FunctionCall != nil {
				toolCalls = append(toolCalls, tools.ToolCall{
					ID:   part.FunctionCall.Name,
					Name: part.FunctionCall.Name,
					Args: part.FunctionCall.Args,
				})
			} else if part.Text != "" {
				contentBuilder.WriteString(part.Text)
			}
		}

		return &tools.LLMResponse{
			Content:   contentBuilder.String(),
			ToolCalls: toolCalls,
		}, nil
	}

	return nil, fmt.Errorf("gemini rate limit exceeded after retry: %w", lastErr)
}

func parseSchema(m map[string]any) *genai.Schema {
	if m == nil {
		return nil
	}
	s := &genai.Schema{}
	if t, ok := m["type"].(string); ok {
		switch t {
		case "object":
			s.Type = genai.TypeObject
		case "string":
			s.Type = genai.TypeString
		case "integer":
			s.Type = genai.TypeInteger
		case "array":
			s.Type = genai.TypeArray
		case "boolean":
			s.Type = genai.TypeBoolean
		}
	}
	if props, ok := m["properties"].(map[string]any); ok {
		s.Properties = make(map[string]*genai.Schema)
		for k, v := range props {
			if vm, ok := v.(map[string]any); ok {
				s.Properties[k] = parseSchema(vm)
			}
		}
	}
	if req, ok := m["required"].([]string); ok {
		s.Required = req
	}
	if desc, ok := m["description"].(string); ok {
		s.Description = desc
	}
	return s
}
