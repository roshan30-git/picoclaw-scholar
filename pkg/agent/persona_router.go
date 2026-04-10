package agent

import (
	"log"
	"strings"

	"github.com/roshan30-git/picoclaw-scholar/pkg/bus"
)

// PersonaType represents the active persona overlay
type PersonaType string

const (
	PersonaNone      PersonaType = "none"
	PersonaIndexer   PersonaType = "agent_indexer"
	PersonaDrill     PersonaType = "agent_drill"
	PersonaExplainer PersonaType = "agent_explainer"
	PersonaScheduler PersonaType = "agent_scheduler"
)

// PersonaRouter uses a zero-API keyword classification to assign a role to each message
type PersonaRouter struct{}

func NewPersonaRouter() *PersonaRouter {
	return &PersonaRouter{}
}

// Route Message maps an incoming user query to a Persona using pure regex/keyword heuristics.
func (pr *PersonaRouter) RouteMessage(msg bus.InboundMessage) PersonaType {
	text := strings.ToLower(msg.Content)

	// 1. Scheduler checks (time/deadlines)
	if pr.matchesAny(text, []string{"due", "deadline", "tomorrow", "tonight", "when is", "remind me", "submission", "schedule"}) {
		return PersonaScheduler
	}

	// 2. Drill checks (quizzes)
	if pr.matchesAny(text, []string{"test me", "quiz", "drill", "mcq", "question", "examine"}) {
		return PersonaDrill
	}

	// 3. Indexer checks (ingestion)
	// Usually triggered by attachments too. For text:
	if pr.matchesAny(text, []string{"index this", "save this", "notes on", "add this to db", "remember this"}) {
		return PersonaIndexer
	}

	// 4. Explainer checks (teaching)
	// If the user seems confused or ask "what is", "how does"
	if pr.matchesAny(text, []string{"what is", "how does", "explain", "i don't understand", "clarify", "tell me about", "concept"}) {
		return PersonaExplainer
	}

	// Default fallback (uses base_soul.md gracefully)
	log.Printf("[Router] No clear persona intended, defaulting to Explainer")
	return PersonaExplainer
}

func (pr *PersonaRouter) matchesAny(text string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}
