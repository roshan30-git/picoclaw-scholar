package sync

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/roshan30-git/picoclaw-scholar/integrations/gdrive"
)

// DriveSyncer watches the local media folder and uploads un-uploaded files to Drive.
type DriveSyncer struct {
	client     *gdrive.Client
	mediaDir   string
	syncRecord string // File that tracks uploaded files
}

func NewDriveSyncer(gc *gdrive.Client) *DriveSyncer {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".studyclaw", "media")
	os.MkdirAll(dir, 0755)

	return &DriveSyncer{
		client:     gc,
		mediaDir:   dir,
		syncRecord: filepath.Join(home, ".studyclaw", "drive_sync_record.txt"),
	}
}

func (d *DriveSyncer) isUploaded(filename string) bool {
	data, err := os.ReadFile(d.syncRecord)
	if err != nil {
		return false
	}
	// Simple check, reading whole file is okay for small scale
	return string(data) != "" && filepath.Base(filename) != "" && contains(string(data), filename)
}

func contains(data, subs string) bool {
	// naive substring
	for i := 0; i <= len(data)-len(subs); i++ {
		if data[i:i+len(subs)] == subs {
			return true
		}
	}
	return false
}

func (d *DriveSyncer) markUploaded(filename string) {
	f, err := os.OpenFile(d.syncRecord, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		f.WriteString(filename + "\n")
	}
}

// Start begins a background loop to scan the media folder periodically
func (d *DriveSyncer) Start(ctx context.Context) {
	log.Println("Drive Syncer started watching ~/.studyclaw/media/")
	ticker := time.NewTicker(2 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				entries, err := os.ReadDir(d.mediaDir)
				if err != nil {
					continue
				}

				for _, e := range entries {
					if e.IsDir() {
						continue
					}

					fname := e.Name()
					if d.isUploaded(fname) {
						continue
					}

					fullPath := filepath.Join(d.mediaDir, fname)
					log.Printf("Syncing %s to Drive...", fname)
					_, err := d.client.UploadFile(ctx, fullPath)
					if err != nil {
						log.Printf("Failed to sync %s: %v", fname, err)
					} else {
						d.markUploaded(fname)
					}
				}
			}
		}
	}()
}
