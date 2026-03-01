package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExcelTool generates Excel (.xlsx) files from study data on request.
type ExcelTool struct {
	exportDir string
}

func NewExcelTool() *ExcelTool {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".studyclaw", "exports")
	os.MkdirAll(dir, 0755)
	return &ExcelTool{exportDir: dir}
}

func (t *ExcelTool) Name() string { return "export_excel" }

func (t *ExcelTool) Description() string {
	return "Generate and export data as an Excel (.xlsx) file. Use when the user asks for a spreadsheet, table export, or Excel report. Accepts structured data as rows."
}

func (t *ExcelTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title": map[string]any{
				"type":        "string",
				"description": "Title/heading for the Excel sheet",
			},
			"headers": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Column header names",
			},
			"rows": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "array"},
				"description": "2D array of cell values, each inner array is a row",
			},
		},
		"required": []string{"title", "headers", "rows"},
	}
}

func (t *ExcelTool) Execute(ctx context.Context, params map[string]any) *ToolResult {
	title, _ := params["title"].(string)
	if title == "" {
		title = "StudyClaw Export"
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"

	// Write headers in bold
	headers, _ := params["headers"].([]any)
	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Write data rows
	rows, _ := params["rows"].([]any)
	for rowIdx, rawRow := range rows {
		row, ok := rawRow.([]any)
		if !ok {
			continue
		}
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	// Save file
	filename := fmt.Sprintf("%s_%s.xlsx", sanitize(title), time.Now().Format("20060102_150405"))
	path := filepath.Join(t.exportDir, filename)
	if err := f.SaveAs(path); err != nil {
		return ErrorResult(fmt.Sprintf("failed to save Excel file: %v", err))
	}

	return SuccessResult(
		fmt.Sprintf("Excel file saved: %s", path),
		fmt.Sprintf("📊 Excel file ready: `%s`\n\nOpen it from your file manager.", filename),
	)
}

func sanitize(s string) string {
	result := make([]byte, 0, len(s))
	for _, c := range []byte(s) {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result = append(result, c)
		} else {
			result = append(result, '_')
		}
	}
	return string(result)
}
