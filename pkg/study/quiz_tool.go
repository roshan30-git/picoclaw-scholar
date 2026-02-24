package study

import (
	"context"
	"fmt"

	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
)

type QuizTool struct {
	engine *QuizEngine
}

func NewQuizTool(engine *QuizEngine) *QuizTool {
	return &QuizTool{engine: engine}
}

func (t *QuizTool) Name() string        { return "generate_quiz" }
func (t *QuizTool) Description() string { return "Generate a multiple-choice quiz on a topic using the student's notes." }

func (t *QuizTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"topic": map[string]any{
				"type":        "string",
				"description": "The topic to generate a quiz for",
			},
			"num_questions": map[string]any{
				"type":        "integer",
				"description": "Number of questions (default 5)",
			},
		},
		"required": []string{"topic"},
	}
}

func (t *QuizTool) Execute(ctx context.Context, params map[string]any) *tools.ToolResult {
	topic, ok := params["topic"].(string)
	if !ok || topic == "" {
		return tools.ErrorResult("topic parameter is required")
	}

	numQ := 5
	if n, ok := params["num_questions"].(float64); ok && n > 0 {
		numQ = int(n)
	}

	quiz, err := t.engine.GenerateQuiz(ctx, topic, numQ)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Quiz generation failed: %v", err))
	}

	return tools.SuccessResult(quiz, fmt.Sprintf("🎯 Quiz on '%s' generated!", topic))
}
