package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Truncate returns a truncated string with ... if it exceeds limit.
func Truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	if limit < 3 {
		return s[:limit]
	}
	return s[:limit-3] + "..."
}

// DownloadOptions specifies options for DownloadFile.
type DownloadOptions struct {
	LoggerPrefix string
}

// DownloadFile downloads a file from URL to local disk.
func DownloadFile(url, filename string, opts DownloadOptions) string {
	// Ensure temp directory exists
	tmpDir := filepath.Join(os.TempDir(), "studyclaw_downloads")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return ""
	}

	destPath := filepath.Join(tmpDir, filepath.Base(filename))

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("[%s] Download failed: %v\n", opts.LoggerPrefix, err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[%s] Download failed status: %d\n", opts.LoggerPrefix, resp.StatusCode)
		return ""
	}

	out, err := os.Create(destPath)
	if err != nil {
		fmt.Printf("[%s] Create file failed: %v\n", opts.LoggerPrefix, err)
		return ""
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("[%s] Copy failed: %v\n", opts.LoggerPrefix, err)
		return ""
	}

	return destPath
}
