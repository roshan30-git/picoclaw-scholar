package study

import (
	"context"
	"fmt"

	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
)

type IngestTool struct {
	engine *IngestionEngine
}

func NewIngestTool(engine *IngestionEngine) *IngestTool {
	return &IngestTool{engine: engine}
}

func (t *IngestTool) Name() string        { return "ingest_pdf" }
func (t *IngestTool) Description() string { return "Extract text from a PDF file and save it to the knowledge base." }

func (t *IngestTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"file_path": map[string]any{
				"type":        "string",
				"description": "Absolute or relative path to the PDF file",
			},
		},
		"required": []string{"file_path"},
	}
}

func (t *IngestTool) Execute(ctx context.Context, params map[string]any) *tools.ToolResult {
	filePath, ok := params["file_path"].(string)
	if !ok || filePath == "" {
		return tools.ErrorResult("file_path parameter is required")
	}

	extracted, err := t.engine.ProcessPDF(filePath)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to process PDF: %v", err))
	}

	preview := extracted
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}

	return tools.SuccessResult(
		fmt.Sprintf("PDF ingested successfully. Extracted %d characters. Preview:\n%s", len(extracted), preview),
		fmt.Sprintf("📥 PDF ingested: %s (%d chars extracted)", filePath, len(extracted)),
	)
}
