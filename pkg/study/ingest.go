package study

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/roshan30-git/picoclaw-scholar/pkg/database"
)

type IngestionEngine struct {
	db *database.DB
}

func NewIngestionEngine(db *database.DB) *IngestionEngine {
	return &IngestionEngine{db: db}
}

// ProcessPDF extracts text from a PDF file and saves it to the database.
func (e *IngestionEngine) ProcessPDF(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}
	defer f.Close()

	// MVP: read raw bytes and do a basic text extraction attempt
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read pdf: %w", err)
	}

	// Simple heuristic text extraction from PDF stream
	content := string(data)
	extracted := extractTextFromPDFContent(content)

	if extracted == "" {
		extracted = "[PDF text extraction returned empty — binary or scanned PDF. OCR needed.]"
	}

	topic := strings.TrimSuffix(strings.Replace(filePath, "\\", "/", -1), ".pdf")
	parts := strings.Split(topic, "/")
	topic = parts[len(parts)-1]

	if err := e.db.SaveNote(topic, extracted, filePath); err != nil {
		log.Printf("Failed to save note: %v", err)
		return extracted, err
	}

	return extracted, nil
}

func extractTextFromPDFContent(raw string) string {
	// Very basic: look for text between BT and ET markers in PDF stream
	var result strings.Builder
	inText := false
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "BT" {
			inText = true
			continue
		}
		if trimmed == "ET" {
			inText = false
			continue
		}
		if inText && strings.Contains(trimmed, "Tj") {
			// Extract string between parentheses
			start := strings.Index(trimmed, "(")
			end := strings.LastIndex(trimmed, ")")
			if start >= 0 && end > start {
				result.WriteString(trimmed[start+1:end] + " ")
			}
		}
	}
	return strings.TrimSpace(result.String())
}
