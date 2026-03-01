package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// WebSearchTool uses the DuckDuckGo Instant Answers API for real-time info.
// No API key required.
type WebSearchTool struct{}

func NewWebSearchTool() *WebSearchTool { return &WebSearchTool{} }

func (t *WebSearchTool) Name() string { return "search_internet" }

func (t *WebSearchTool) Description() string {
	return "Search the internet for current events, news, or real-time factual information. Use when asked about today's events, recent news, or anything requiring up-to-date data."
}

func (t *WebSearchTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The search query to look up",
			},
		},
		"required": []string{"query"},
	}
}

func (t *WebSearchTool) Execute(ctx context.Context, params map[string]any) *ToolResult {
	query, ok := params["query"].(string)
	if !ok || strings.TrimSpace(query) == "" {
		return ErrorResult("query parameter is required")
	}

	result, err := duckDuckGoSearch(ctx, query)
	if err != nil {
		return ErrorResult(fmt.Sprintf("search failed: %v", err))
	}

	return SuccessResult(result, result)
}

type ddgResponse struct {
	AbstractText  string `json:"AbstractText"`
	AbstractURL   string `json:"AbstractURL"`
	RelatedTopics []struct {
		Text string `json:"Text"`
	} `json:"RelatedTopics"`
}

func duckDuckGoSearch(ctx context.Context, query string) (string, error) {
	apiURL := "https://api.duckduckgo.com/?q=" + url.QueryEscape(query) + "&format=json&no_redirect=1&no_html=1"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "StudyClaw/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ddg ddgResponse
	if err := json.Unmarshal(body, &ddg); err != nil {
		return "", err
	}

	var parts []string
	if ddg.AbstractText != "" {
		parts = append(parts, ddg.AbstractText)
		if ddg.AbstractURL != "" {
			parts = append(parts, "Source: "+ddg.AbstractURL)
		}
	}

	for i, topic := range ddg.RelatedTopics {
		if i >= 3 {
			break
		}
		if topic.Text != "" {
			parts = append(parts, "• "+topic.Text)
		}
	}

	if len(parts) == 0 {
		return "No results found for: " + query, nil
	}

	return strings.Join(parts, "\n"), nil
}
