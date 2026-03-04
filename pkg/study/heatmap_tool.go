package study

import (
	"context"

	"github.com/roshan30-git/picoclaw-scholar/pkg/database"
	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
)

// HeatmapTool renders the weak topic heatmap for the LLM.
type HeatmapTool struct {
	db *database.DB
}

func NewHeatmapTool(db *database.DB) *HeatmapTool {
	return &HeatmapTool{db: db}
}

func (t *HeatmapTool) Name() string { return "view_heatmap" }
func (t *HeatmapTool) Description() string {
	return "Show the student's weak topic heatmap with color-coded scores and progress bars."
}

func (t *HeatmapTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (t *HeatmapTool) Execute(ctx context.Context, params map[string]any) *tools.ToolResult {
	heatmap := FormatWeakTopicHeatmap(t.db)
	progress := FormatProgressBars(t.db)
	return tools.SuccessResult(heatmap+"\n\n"+progress, "Generated heatmap and progress bars.")
}
