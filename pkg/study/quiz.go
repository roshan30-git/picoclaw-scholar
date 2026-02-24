package study

import (
	"context"
	"fmt"
	"os"

	"github.com/user/studyclaw/pkg/database"
	"github.com/user/studyclaw/pkg/tools"
)

type QuizEngine struct {
	provider tools.LLMProvider
	db       *database.DB
}

func NewQuizEngine(provider tools.LLMProvider, db *database.DB) *QuizEngine {
	return &QuizEngine{provider: provider, db: db}
}

// GenerateQuiz creates a quiz for the given topic using DB context and the Drill Sergeant persona.
func (q *QuizEngine) GenerateQuiz(ctx context.Context, topic string, numQuestions int) (string, error) {
	dbContext, _ := q.db.QueryContext(topic)

	persona, _ := os.ReadFile("workspace/PROMPTS/drill_sergeant.txt")
	if persona == nil {
		persona = []byte("You are a strict quiz master.")
	}

	prompt := fmt.Sprintf(`%s

Context from student's notes:
---
%s
---

Generate %d multiple-choice questions on the topic "%s".
Format as JSON array: [{"q":"...","options":["A)...","B)...","C)...","D)..."],"answer":"A"}]
Only output the JSON, nothing else.`, string(persona), dbContext, numQuestions, topic)

	messages := []tools.Message{
		{Role: "user", Content: prompt},
	}

	resp, err := q.provider.Chat(ctx, messages, nil, "", nil)
	if err != nil {
		return "", fmt.Errorf("quiz generation failed: %w", err)
	}

	return resp.Content, nil
}
