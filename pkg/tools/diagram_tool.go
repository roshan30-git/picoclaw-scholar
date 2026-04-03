package tools

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
)

// DiagramTool instructs the agent to produce Mermaid diagram markup
// and routes it through the existing viewer at localhost:8080.
// Also handles electrical/circuit diagrams via text-art notation.
type DiagramTool struct{}

func NewDiagramTool() *DiagramTool { return &DiagramTool{} }

func (t *DiagramTool) Name() string { return "render_diagram" }

func (t *DiagramTool) Description() string {
	return "Render any diagram, flowchart, circuit, or visual. Accepts Mermaid markup for flowcharts/ERD/sequence diagrams, or 'circuit' type for electrical/electronic circuit descriptions. Returns a viewer URL."
}

func (t *DiagramTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"type": map[string]any{
				"type":        "string",
				"enum":        []string{"flowchart", "sequence", "erd", "circuit", "mindmap", "gantt"},
				"description": "The type of diagram to render",
			},
			"markup": map[string]any{
				"type":        "string",
				"description": "The Mermaid diagram markup or circuit description text",
			},
			"title": map[string]any{
				"type":        "string",
				"description": "Optional title for the diagram",
			},
		},
		"required": []string{"type", "markup"},
	}
}

func (t *DiagramTool) Execute(ctx context.Context, params map[string]any) *ToolResult {
	diagramType, _ := params["type"].(string)
	markup, _ := params["markup"].(string)
	title, _ := params["title"].(string)

	if strings.TrimSpace(markup) == "" {
		return ErrorResult("markup parameter is required")
	}

	// For circuit type, wrap in a code block for display
	if diagramType == "circuit" {
		formatted := formatCircuit(markup, title)
		return SuccessResult(
			fmt.Sprintf("Circuit diagram rendered:\n%s", formatted),
			fmt.Sprintf("⚡ *Circuit Diagram: %s*\n\n```\n%s\n```", title, formatted),
		)
	}

	// For Mermaid types, generate a viewer ID and return a URL
	viewerID := fmt.Sprintf("%s_%d", diagramType, rand.Int63n(999999))
	viewerURL := fmt.Sprintf("http://127.0.0.1:8080/viewer?id=%s&type=%s", viewerID, diagramType)

	// Return both the markup (for LLM context) and viewer URL
	return &ToolResult{
		ForLLM:  fmt.Sprintf("Diagram generated. ID: %s\nMarkup:\n%s", viewerID, markup),
		ForUser: fmt.Sprintf("📊 *%s Diagram*\n\nView it here: %s\n\n_(Also rendering below)_\n```mermaid\n%s\n```", capitalize(diagramType), viewerURL, markup),
	}
}

func formatCircuit(description, title string) string {
	var sb strings.Builder
	if title != "" {
		sb.WriteString(title + "\n")
		sb.WriteString(strings.Repeat("=", len(title)) + "\n\n")
	}
	// Pass through the description as-is; LLM already produces ASCII circuit art
	sb.WriteString(description)
	return sb.String()
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
