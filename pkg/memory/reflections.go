package memory

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReflectionManager handles storing and retrieving the AI's "lessons learned"
// to prevent it from repeating mistakes across persistent sessions.
type ReflectionManager struct {
	baseDir string
}

func NewReflectionManager(workspacePath string) *ReflectionManager {
	dir := filepath.Join(workspacePath, "MEMORY", "reflections")
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("[Memory] Warning: Could not create reflections directory: %v", err)
	}
	return &ReflectionManager{baseDir: dir}
}

// LogMistake writes a new lesson learned to a daily markdown file.
// Triggered when the user says "wrong", "actually", or explicitly corrects the AI.
func (r *ReflectionManager) LogMistake(userCorrection string, aiOriginalContext string) error {
	dateStr := time.Now().Format("2006-01-02")
	filePath := filepath.Join(r.baseDir, fmt.Sprintf("%s.md", dateStr))

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	timestamp := time.Now().Format("15:04:05")
	entry := fmt.Sprintf("\n## Lesson Learned at %s\n**Context:** %s\n**Correction:** %s\n---\n", timestamp, aiOriginalContext, userCorrection)

	_, err = f.WriteString(entry)
	if err == nil {
		log.Println("[Memory] 🧠 New reflection logged: AI learned from a mistake.")
	}
	return err
}

// GetRecentReflections pulls the contents of the last 3 days of reflections to inject into the system prompt.
func (r *ReflectionManager) GetRecentReflections() string {
	var lessons []string
	
	// Scan the directory for .md files
	entries, err := os.ReadDir(r.baseDir)
	if err != nil || len(entries) == 0 {
		return ""
	}

	// In a complete system, we'd sort by date and pick the last 3.
	// For the MVP, we just read all existing reflections (assuming low volume).
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			path := filepath.Join(r.baseDir, e.Name())
			file, err := os.Open(path)
			if err != nil {
				continue
			}
			content, _ := io.ReadAll(file)
			lessons = append(lessons, string(content))
			file.Close()
		}
	}

	if len(lessons) == 0 {
		return ""
	}

	return "🧠 CRITICAL SYSTEM MEMORY (Do not repeat these past mistakes):\n" + strings.Join(lessons, "\n")
}
