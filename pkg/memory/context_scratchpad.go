package memory

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/roshan30-git/picoclaw-scholar/pkg/database"
	"github.com/roshan30-git/picoclaw-scholar/pkg/study"
	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
)

// AcademicProfile merges with StudentProfile
type AcademicProfile struct {
	Semester       string
	University     string
	HighYieldTopic string
}

type TemporalContext struct {
	CurrentDate   time.Time
	DaysUntilExam int
	FestivalMode  bool
}

type SessionContext struct {
	Academic   AcademicProfile
	Temporal   TemporalContext
	ShortTerm  string // Summary of last 2 exchanges
	Background string // Top 2 semantic search results from DB
}

type ContextManager struct {
	db          *database.DB
	provider    tools.LLMProvider
	profileMgr  *ProfileManager
	deadlineMgr *study.DeadlineTracker
}

func NewContextManager(db *database.DB, provider tools.LLMProvider, profileMgr *ProfileManager, tracker *study.DeadlineTracker) *ContextManager {
	return &ContextManager{
		db:          db,
		provider:    provider,
		profileMgr:  profileMgr,
		deadlineMgr: tracker,
	}
}

func (cm *ContextManager) BuildPrompt(ctx context.Context, chatID string, userMessage string, shortTermSummary string) string {
	var builder strings.Builder

	cm.injectProfile(&builder)
	cm.injectTemporalContext(&builder)
	cm.injectShortTermMemory(&builder, userMessage, shortTermSummary)
	cm.injectBackgroundKnowledge(&builder, userMessage)

	return builder.String()
}

func (cm *ContextManager) injectProfile(builder *strings.Builder) {
	if cm.profileMgr == nil {
		return
	}
	if profile := cm.profileMgr.GetProfile(); profile != nil {
		builder.WriteString("### ACADEMIC CONTEXT ###\n")
		builder.WriteString(profile.FormatForPrompt() + "\n")
	}
}

func (cm *ContextManager) injectTemporalContext(builder *strings.Builder) {
	builder.WriteString("### TEMPORAL CONTEXT ###\n")
	now := time.Now()
	builder.WriteString(fmt.Sprintf("Current Date: %s\n", now.Format(time.RFC822)))

	if cm.deadlineMgr == nil {
		return
	}

	deadlines, _ := cm.deadlineMgr.GetUpcoming()
	if len(deadlines) == 0 {
		return
	}

	closest := deadlines[0]
	daysUntil := int(time.Until(closest.DueDate).Hours() / 24)
	builder.WriteString(fmt.Sprintf("Days until next major deadline (%s): %d days\n", closest.Title, daysUntil))

	if daysUntil < 14 {
		builder.WriteString("⚠️ EXAM INTENSITY MODE ACTIVE: Provide high-yield, concise, rigorous answers aimed at rapid revision.\n")
	}
}

func (cm *ContextManager) injectShortTermMemory(builder *strings.Builder, userMessage, shortTermSummary string) {
	if shortTermSummary == "" {
		return
	}

	if hasPronouns(userMessage) {
		builder.WriteString("### SHORT TERM SCRATCHPAD (Recent Chat Context) ###\n")
		builder.WriteString(shortTermSummary + "\n")
	}
}

func (cm *ContextManager) injectBackgroundKnowledge(builder *strings.Builder, userMessage string) {
	backgroundNotes, _ := cm.db.QueryContext(userMessage)
	if backgroundNotes == "" {
		return
	}

	builder.WriteString("### BACKGROUND KNOWLEDGE (Semantic Hits) ###\n")
	words := strings.Fields(backgroundNotes)
	if len(words) > 600 {
		backgroundNotes = strings.Join(words[:600], " ") + "..."
	}
	builder.WriteString(backgroundNotes + "\n")
}

func hasPronouns(msg string) bool {
	pronouns := []string{" it", " that", " his", " this", " she", " he", " they", " them"}
	lower := strings.ToLower(msg)
	for _, p := range pronouns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

// SummarizeAndClear fires after every 3rd message to roll the context
func (cm *ContextManager) SummarizeAndClear(ctx context.Context, chatID string, history []tools.Message, callback func(newSummary string)) {
	go func() {
		log.Printf("Triggering rolling summary for chat %s", chatID)

		var chatLog strings.Builder
		for _, msg := range history {
			// Skip system prompts to save tokens
			if msg.Role == "system" {
				continue
			}
			chatLog.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}

		prompt := fmt.Sprintf("Summarize the following conversation in 50 words or less, retaining all critical facts, variables, or entities mentioned:\n\n%s", chatLog.String())

		// Call LLM for generation
		resp, err := cm.provider.Chat(ctx, []tools.Message{{Role: "user", Content: prompt}}, nil, "agent_explainer", nil)
		if err != nil {
			log.Printf("Failed to generate short-term summary: %v", err)
			return
		}

		// Push exactly to database
		err = cm.db.SaveChatSummary(chatID, resp.Content)
		if err != nil {
			log.Printf("Failed to save short-term summary: %v", err)
		}

		// Signal the callback that summary is complete, so loop can clear raw history
		callback(resp.Content)
	}()
}

// GetLatestChatSummary pulls the rolling summary for the chat session
func (cm *ContextManager) GetLatestChatSummary(chatID string) string {
	return cm.db.GetLatestChatSummary(chatID)
}
