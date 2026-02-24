package providers

import (
	"context"
	"fmt"
	"strings"

	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"

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

	if len(toolDefs) > 0 {
		var decls []*genai.FunctionDeclaration
		for _, td := range toolDefs {
			decls = append(decls, &genai.FunctionDeclaration{
				Name:        td.Function.Name,
				Description: td.Function.Description,
				Parameters:  parseSchema(td.Function.Parameters),
			})
		}
		m.Tools = []*genai.Tool{{FunctionDeclarations: decls}}
	}

	var parts []genai.Part
	for _, msg := range messages {
		if msg.Role == "model" {
			parts = append(parts, genai.Text(msg.Content)) // For tool results/model history
		} else if msg.Role == "user" {
			parts = append(parts, genai.Text(msg.Content))
		}
	}

	resp, err := m.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, fmt.Errorf("gemini generate: %w", err)
	}

	var contentBuilder strings.Builder
	var toolCalls []tools.ToolCall
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if funcCall, ok := part.(genai.FunctionCall); ok {
					toolCalls = append(toolCalls, tools.ToolCall{
						ID:   funcCall.Name, // Using name as ID for simple implementations
						Name: funcCall.Name,
						Args: funcCall.Args,
					})
				} else if text, ok := part.(genai.Text); ok {
					contentBuilder.WriteString(string(text))
				}
			}
		}
	}

	return &tools.LLMResponse{
		Content:   contentBuilder.String(),
		ToolCalls: toolCalls,
	}, nil
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
		s.Required = req // Note: In newer genai, Required is []string
	}
	if desc, ok := m["description"].(string); ok {
		s.Description = desc
	}
	return s
}
