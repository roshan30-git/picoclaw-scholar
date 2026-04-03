package study

import (
	"context"
	"fmt"
	"log"

	"github.com/roshan30-git/picoclaw-scholar/pkg/bus"
	"github.com/roshan30-git/picoclaw-scholar/pkg/database"
	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
)

type WeeklyCardsGenerator struct {
	db       *database.DB
	provider tools.LLMProvider
	bus      *bus.MessageBus
	ownerID  string
}

func NewWeeklyCardsGenerator(db *database.DB, provider tools.LLMProvider, b *bus.MessageBus, ownerEnv string) *WeeklyCardsGenerator {
	return &WeeklyCardsGenerator{
		db:       db,
		provider: provider,
		bus:      b,
		ownerID:  ownerEnv,
	}
}

func (w *WeeklyCardsGenerator) GenerateAndSend(ctx context.Context) {
	log.Println("[WeeklyCards] Triggering Sunday flashcards...")

	// Query weak topics where avg_score < 70 (Assuming 100 point total mapping later)
	// Right now, if they did poorly on drills, they get added. For MVP, we fetch recent topics to review.
	// We'll mimic fetching weak topics by just pulling recent distinct topics
	rows, err := w.db.Conn().Query(`SELECT DISTINCT topic FROM notes ORDER BY created_at DESC LIMIT 3`)
	if err != nil {
		log.Printf("[WeeklyCards] Failed to fetch topics: %v", err)
		return
	}
	defer rows.Close()

	var topics []string
	for rows.Next() {
		var t string
		rows.Scan(&t)
		topics = append(topics, t)
	}

	if len(topics) == 0 {
		return
	}

	prompt := fmt.Sprintf("Generate 5 concise study flashcards for the following topics: %v\nFormat strictly as a Markdown checklist or bullet list. Keep them highly informative and focused on key facts or formulas.", topics)

	msg := []tools.Message{{Role: "user", Content: prompt}}
	resp, err := w.provider.Chat(ctx, msg, nil, "gemini-2.0-flash", nil)
	if err != nil {
		log.Printf("[WeeklyCards] LLM error: %v", err)
		return
	}

	// Add the Formula render button to make it interactive via visual interceptor logic in agent loop later,
	// but here we just send standard text
	out := bus.OutboundMessage{
		ChatID:  w.ownerID,
		Content: "🗂️ *__Your Sunday Revision Cards__*\n\n" + resp.Content,
		Channel: "whatsapp",
	}
	w.bus.PublishOutbound(out)
}
