package study

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/genai-go"
	"github.com/roshan30-git/picoclaw-scholar/pkg/database"
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

	f, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("open image: %w", err)
	}
	defer f.Close()

	// In a real pipeline, we'd add local OpenCV contrast/deskew here before sending 
	// to save tokens, but Gemini 2.0 Flash is robust enough for Phase 1.

	uploadRes, err := o.client.Files.UploadFile(ctx, filepath.Base(imagePath), f, &genai.UploadFileOptions{})
	if err != nil {
		return "", fmt.Errorf("upload to gemini: %w", err)
	}
	defer o.client.Files.DeleteFile(ctx, uploadRes.Name)

	prompt := "Extract all handwritten notes, text, and mathematical formulas from this image. Output ONLY the extracted text formatted cleanly in Markdown. If there are code snippets or circuits, describe them."

	resp, err := o.client.Models.GenerateContent(
		ctx,
		o.model,
		genai.FileData{URI: uploadRes.URI},
		genai.Text(prompt),
	)
	if err != nil {
		return "", fmt.Errorf("generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no text extracted")
	}

	extractedText := ""
	if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		extractedText = string(textPart)
	}

	title := fmt.Sprintf("OCR Note - %s", filepath.Base(imagePath))
	if err := o.db.SaveNote(title, extractedText, "whatsapp_ocr"); err != nil {
		return extractedText, fmt.Errorf("save to db: %w", err)
	}

	log.Printf("[OCR] Successfully indexed %d characters", len(extractedText))
	return extractedText, nil
}
