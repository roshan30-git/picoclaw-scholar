package study

import (
	"context"
	"fmt"

	"github.com/roshan30-git/picoclaw-scholar/pkg/database"
	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
)

// SearchNotesTool allows the LLM to search the student's ingested PDF notes via SQLite full-text search.
type SearchNotesTool struct {
	db *database.DB
}

func NewSearchNotesTool(db *database.DB) *SearchNotesTool {
	return &SearchNotesTool{db: db}
}

func (t *SearchNotesTool) Name() string {
	return "search_notes"
}

func (t *SearchNotesTool) Description() string {
	return "Search the knowledge base for information from the student's ingested PDFs and notes. Use this to answer questions about specific topics they have studied."
}

func (t *SearchNotesTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The topic, keyword, or concept to search for in the notes.",
			},
		},
		"required": []string{"query"},
	}
}

func (t *SearchNotesTool) Execute(ctx context.Context, params map[string]any) *tools.ToolResult {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return tools.ErrorResult("query parameter is required")
	}

	results, err := t.db.QueryContext(query)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Database search failed: %v", err))
	}

	if results == "" {
		return tools.SuccessResult("No matching notes found for that query.", "No notes found.")
	}

	msg := fmt.Sprintf("🔍 Found notes for '%s'. Retrieved context:\n\n%s", query, results)
	return tools.SuccessResult(msg, fmt.Sprintf("Retrieved notes for '%s'.", query))
}
