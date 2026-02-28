package providers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/roshan30-git/picoclaw-scholar/pkg/bus"
)

// TokenTracker monitors daily token usage and alerts the owner if limits are approached.
// Free tier has 1,500 requests per day (1M tokens).
type TokenTracker struct {
	mu            sync.Mutex
	dailyTokens   int
	dailyReqs     int
	lastResetDate string
	dailyLimit    int
	bus           *bus.MessageBus
	ownerID       string
	alertSent     bool
}

func NewTokenTracker(b *bus.MessageBus, ownerID string) *TokenTracker {
	limitStr := os.Getenv("STUDYCLAW_DAILY_TOKEN_LIMIT")
	limit := 1000000 // default 1M
	if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
		limit = v
	}

	return &TokenTracker{
		lastResetDate: time.Now().Format("2006-01-02"),
		dailyLimit:    limit,
		bus:           b,
		ownerID:       ownerID,
	}
}

func (t *TokenTracker) checkReset() {
	today := time.Now().Format("2006-01-02")
	if today != t.lastResetDate {
		t.dailyTokens = 0
		t.dailyReqs = 0
		t.lastResetDate = today
		t.alertSent = false
		log.Println("[TokenTracker] Rolled over daily token count.")
	}
}

// AddUsage logs the tokens used in a request and checks thresholds.
func (t *TokenTracker) AddUsage(ctx context.Context, inputTokens, outputTokens int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.checkReset()

	total := inputTokens + outputTokens
	t.dailyTokens += total
	t.dailyReqs++

	// Alert owner if we cross 80% usage
	threshold := int(float64(t.dailyLimit) * 0.80)
	if t.dailyTokens > threshold && !t.alertSent {
		t.alertSent = true
		msg := fmt.Sprintf("⚠️ *__Cost Alert__*\n\nTokens consumed today: %d / %d (%.1f%%).\nNearing Gemini free-tier daily limits.", 
			t.dailyTokens, t.dailyLimit, float64(t.dailyTokens)/float64(t.dailyLimit)*100.0)
		
		log.Println(msg)
		
		if t.bus != nil && t.ownerID != "" {
			t.bus.Publish(bus.OutboundMessage{
				ChatID:  t.ownerID,
				Content: msg,
				Channel: "whatsapp", // default to whatsapp alert
			})
		}
	}
}
