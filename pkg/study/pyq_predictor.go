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

// PredictQuestions uses gemini-2.5-pro to analyze patterns in the past 5 years of exam papers.
func (p *PYQPredictor) PredictQuestions(ctx context.Context, subject string) (string, error) {
	prompt := fmt.Sprintf(`As an Exam Strategy AI, analyze the structural frequency of past 5 years papers for: '%s'.
Focus on the current syllabus format. Predict the top 5 most likely conceptual topics.
Provide a probability percentage and a brief justification for each. Keep the output concise.`, subject)

	msg := []tools.Message{{Role: "user", Content: prompt}}
	
	// Force-route reasoning task to the Pro model
	resp, err := p.provider.Chat(ctx, msg, nil, "gemini-2.5-pro", nil)
	if err != nil {
		return "", fmt.Errorf("prediction failed: %w", err)
	}

	return resp.Content, nil
}
