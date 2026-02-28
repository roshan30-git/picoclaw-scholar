package tools

import (
    "context"
    "fmt"
)

type PYQPredictTool struct {
    Predictor interface {
        PredictQuestions(ctx context.Context, subject string) (string, error)
    }
}

func NewPYQPredictTool(p interface{ PredictQuestions(context.Context, string)(string, error) }) *PYQPredictTool {
    return &PYQPredictTool{Predictor: p}
}

func (t *PYQPredictTool) Name() string { return "predict_questions" }

func (t *PYQPredictTool) Description() string {
    return "Predicts the most likely exam questions and topics for a given subject based on historical PYQ (Past Year Question) pattern analysis."
}

func (t *PYQPredictTool) Parameters() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "subject": map[string]interface{}{
                "type":        "string",
                "description": "The name of the subject or course (e.g., 'Circuit Theory', 'Data Structures')",
            },
        },
        "required": []string{"subject"},
    }
}

func (t *PYQPredictTool) Execute(ctx context.Context, args map[string]interface{}) ToolResult {
    subject, _ := args["subject"].(string)
    if subject == "" {
        return ToolResult{ForLLM: "Error: subject is required"}
    }
    
    predictions, err := t.Predictor.PredictQuestions(ctx, subject)
    if err != nil {
        return ToolResult{ForLLM: fmt.Sprintf("Error predicting questions: %v", err)}
    }
    
    return ToolResult{ForLLM: predictions}
}

type MockPaperTool struct {
    LLMProvider LLMProvider
}

func NewMockPaperTool(provider LLMProvider) *MockPaperTool {
    return &MockPaperTool{LLMProvider: provider}
}

func (t *MockPaperTool) Name() string { return "generate_mock_paper" }

func (t *MockPaperTool) Description() string {
    return "Generates a complete mock exam paper for a specific subject based on the syllabus and past patterns."
}

func (t *MockPaperTool) Parameters() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "subject": map[string]interface{}{
                "type":        "string",
                "description": "The subject name",
            },
            "difficulty": map[string]interface{}{
                "type": "string",
                "description": "Difficulty level: 'easy', 'medium', or 'hard'",
            },
        },
        "required": []string{"subject"},
    }
}

func (t *MockPaperTool) Execute(ctx context.Context, args map[string]interface{}) ToolResult {
    subject, _ := args["subject"].(string)
    difficulty, _ := args["difficulty"].(string)
    if difficulty == "" {
        difficulty = "medium"
    }

    prompt := fmt.Sprintf("Generate a comprehensive %s difficulty mock exam paper for the subject '%s'. Include 5 varied questions covering different syllabus units.", difficulty, subject)
    
    resp, err := t.LLMProvider.Chat(ctx, []Message{{Role: "user", Content: prompt}}, nil, "gemini-2.0-flash", nil)
    if err != nil {
        return ToolResult{ForLLM: fmt.Sprintf("Failed to generate mock paper: %v", err)}
    }
    
    return ToolResult{ForLLM: fmt.Sprintf("Mock Paper Generated:\n\n%s", resp.Content)}
}
