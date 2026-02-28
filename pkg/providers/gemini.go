package providers

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
	"google.golang.org/genai"
)

// ─────────────────────────────────────────────
//  FREE TIER MODEL SELECTION (updated Feb 28, 2026)
// ─────────────────────────────────────────────
//
//  Model                  | Free API? | RPM | RPD   | Notes
//  ─────────────────────────────────────────────────────────
//  gemini-2.0-flash-lite  |   ✅ YES  |  15 | 1000  | BEST for daily personal bot
//  gemini-2.0-flash       |   ✅ YES  |  10 |  200  | Smarter, fewer requests
//  gemini-2.5-flash       |   ✅ YES  |  10 |  250  | Latest stable model
//  gemini-2.5-flash-lite  |   ✅ YES  |  15 | 1000  | Same quota as 2.0-flash-lite
//  gemini-2.5-pro         |   ✅ YES  |   5 |  100  | Too limited for daily use
//  gemini-3-flash         |   ✅ YES  |  10 |  250  | Available Dec 2025 onward
//  gemini-3.1-flash-*     |   ⚠️ YES  |   - |   -   | Image gen preview only
//  gemini-3-pro           |   ❌ NO   |   - |   -   | No free API tier (app only)
//  gemini-3.1-pro         |   ❌ NO   |   - |   -   | No free API tier (app only)
//  *thinking* models      |   ❌ NO   |   - |   -   | No free API tier for thinking
//
//  Conclusion for StudyClaw:
//  → Use gemini-2.0-flash-lite as default (1000 RPD is perfect for a personal bot)
//  → Use gemini-2.5-flash for quiz generation (slightly smarter, still free)
//  → DO NOT use Pro/Thinking models – not available on the free dev API key

const (
	// ModelChat is used for all conversational replies. Most generous free limits.
	ModelChat = "gemini-2.0-flash-lite"

	// ModelQuiz is used for structured quiz generation where quality matters more.
	ModelQuiz = "gemini-2.5-flash"
)

// ─────────────────────────────────────────────
//  Token-bucket rate limiter (api-patterns/rate-limiting.md)
//  Gemini free tier: 15 RPM for gemini-2.0-flash-lite
//  We limit ourselves to 12 RPM (20% headroom) to avoid 429s.
// ─────────────────────────────────────────────

const maxRPM = 12 // requests per minute budget

type rateLimiter struct {
	mu      sync.Mutex
	tokens  int
	maxTok  int
	lastFil time.Time
}

func newRateLimiter(rpm int) *rateLimiter {
	return &rateLimiter{tokens: rpm, maxTok: rpm, lastFil: time.Now()}
}

// Wait blocks until a token is available.
func (r *rateLimiter) Wait() {
	for {
		r.mu.Lock()
		now := time.Now()
		// Refill at 1 token per (60/maxTok) seconds
		elapsed := now.Sub(r.lastFil)
		refill := int(elapsed.Seconds() / (60.0 / float64(r.maxTok)))
		if refill > 0 {
			r.tokens = min(r.maxTok, r.tokens+refill)
			r.lastFil = now
		}
		if r.tokens > 0 {
			r.tokens--
			r.mu.Unlock()
			return
		}
		r.mu.Unlock()
		time.Sleep(500 * time.Millisecond)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ─────────────────────────────────────────────
//  GeminiProvider
// ─────────────────────────────────────────────

type GeminiProvider struct {
	client  *genai.Client
	model   string
	limiter *rateLimiter
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
	return &GeminiProvider{
		client:  client,
		model:   ModelChat,
		limiter: newRateLimiter(maxRPM),
	}, nil
}

func (g *GeminiProvider) GetDefaultModel() string { return g.model }

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

	// Build content list
	var contents []*genai.Content
	for _, msg := range messages {
		role := "user"
		if msg.Role == "model" {
			role = "model"
		}
		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: []*genai.Part{genai.NewPartFromText(msg.Content)},
		})
	}

	// Build config (tools if requested)
	cfg := &genai.GenerateContentConfig{}
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

	// Apply rate-limit token bucket (api-patterns)
	g.limiter.Wait()

	// Retry once on 429 (rate limit exceeded from service side)
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(10 * time.Second)
		}

		resp, err := g.client.Models.GenerateContent(ctx, model, contents, cfg)
		if err != nil {
			lastErr = err
			if strings.Contains(err.Error(), "429") {
				continue
			}
			return nil, fmt.Errorf("gemini generate: %w", err)
		}

		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			return &tools.LLMResponse{Content: ""}, nil
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

// ─────────────────────────────────────────────
//  Schema helpers
// ─────────────────────────────────────────────

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
