package study

import (
	"context"
	"fmt"
	
	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
)

type PYQPredictor struct {
	provider tools.LLMProvider
}

func NewPYQPredictor(provider tools.LLMProvider) *PYQPredictor {
	return &PYQPredictor{provider: provider}
}

// PredictQuestions uses a reasoning model (gemini-2.5-pro) to analyze past questions and predict likely exam topics.
func (p *PYQPredictor) PredictQuestions(ctx context.Context, subject string) (string, error) {
	prompt := fmt.Sprintf(`As an Exam Strategy AI, analyze the structural frequency of past 5 years papers for the subject: '%s'.
Focus heavily on the current/new syllabus format. Predict the top 5 most likely exam conceptual topics that will appear in the upcoming test. 
Provide a percentage probability for each topic and a brief justification based on recent patterns. Keep it concise to save context window tokens.`, subject)

	msg := []tools.Message{{Role: "user", Content: prompt}}
	
	// Force route this complex reasoning task to the Pro model
	resp, err := p.provider.Chat(ctx, msg, nil, "gemini-2.5-pro", nil)
	if err != nil {
		return "", fmt.Errorf("prediction failed: %w", err)
	}

	return resp.Content, nil
}
