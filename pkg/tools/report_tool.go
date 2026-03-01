package tools

import (
	"context"
	"fmt"
	"time"
)

type ReportGeneratorTool struct {
	LLMProvider LLMProvider
}

func NewReportGeneratorTool(provider LLMProvider) *ReportGeneratorTool {
	return &ReportGeneratorTool{LLMProvider: provider}
}

func (t *ReportGeneratorTool) Name() string { return "generate_report" }

func (t *ReportGeneratorTool) Description() string {
	return "Generates a structured, comprehensive academic markdown report on a specified topic. Outputs a well-formatted document summarizing key facts, theories, and examples."
}

func (t *ReportGeneratorTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"topic": map[string]interface{}{
				"type":        "string",
				"description": "The specific topic to generate a report on.",
			},
			"depth": map[string]interface{}{
				"type":        "string",
				"description": "Depth of the report: 'brief', 'standard', or 'comprehensive'.",
			},
		},
		"required": []string{"topic"},
	}
}

func (t *ReportGeneratorTool) Execute(ctx context.Context, args map[string]interface{}) *ToolResult {
	topic, _ := args["topic"].(string)
	depth, _ := args["depth"].(string)

	if topic == "" {
		return &ToolResult{ForLLM: "Error: topic is required."}
	}
	if depth == "" {
		depth = "standard"
	}

	prompt := fmt.Sprintf("Write a %s academic report on the topic: '%s'. Structure it with an Introduction, Key Concepts, Real-World Applications, and a Conclusion. Use Markdown styling (headers, bullets, bold text).", depth, topic)

	// Since reports are generative and relatively simple structured text, flash is adequate.
	resp, err := t.LLMProvider.Chat(ctx, []Message{{Role: "user", Content: prompt}}, nil, "gemini-2.0-flash", nil)
	if err != nil {
		return &ToolResult{ForLLM: fmt.Sprintf("Failed to generate report: %v", err)}
	}

	// In a full implementation, we might write this to a .md file and send it as a document via WhatsApp.
	// For the MVP, we just return the markdown text back to the loop to send to the user.
	header := fmt.Sprintf("📄 *Automated Report Generated [%s]*\n\n", time.Now().Format("2006-01-02"))
	return &ToolResult{ForLLM: header + resp.Content}
}
