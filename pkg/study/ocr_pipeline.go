package study

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/roshan30-git/picoclaw-scholar/pkg/database"
	"google.golang.org/genai"
)

// OCRPipeline handles extracting text from images using Gemini Vision.
// This is the cost-efficient choice for handwriting vs Tesseract.
type OCRPipeline struct {
	client *genai.Client
	db     *database.DB
	model  string
}

func NewOCRPipeline(apiKey string, db *database.DB) (*OCRPipeline, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, err
	}
	return &OCRPipeline{
		client: client,
		db:     db,
		model:  "gemini-2.0-flash",
	}, nil
}

// ExtractAndSave processes an image file, extracts text/formulas, and saves to the notes DB.
func (o *OCRPipeline) ExtractAndSave(ctx context.Context, imagePath string) (string, error) {
	log.Printf("[OCR] Starting extraction for %s", filepath.Base(imagePath))

	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("read image: %w", err)
	}

	// Detect mime type based on extension
	mimeType := "image/jpeg"
	switch filepath.Ext(imagePath) {
	case ".png":
		mimeType = "image/png"
	case ".webp":
		mimeType = "image/webp"
	case ".gif":
		mimeType = "image/gif"
	}

	prompt := "Extract all handwritten notes, text, and mathematical formulas from this image. Output ONLY the extracted text formatted cleanly in Markdown. If there are code snippets or circuits, describe them."

	resp, err := o.client.Models.GenerateContent(
		ctx,
		o.model,
		[]*genai.Content{
			{
				Parts: []*genai.Part{
					{InlineData: &genai.Blob{MIMEType: mimeType, Data: imageBytes}},
					{Text: prompt},
				},
			},
		},
		&genai.GenerateContentConfig{},
	)
	if err != nil {
		return "", fmt.Errorf("generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return "", fmt.Errorf("no text extracted")
	}

	extractedText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			extractedText += part.Text
		}
	}

	title := fmt.Sprintf("OCR Note - %s", filepath.Base(imagePath))
	if err := o.db.SaveNote(title, extractedText, "whatsapp_ocr"); err != nil {
		return extractedText, fmt.Errorf("save to db: %w", err)
	}

	log.Printf("[OCR] Successfully indexed %d characters", len(extractedText))
	return extractedText, nil
}
