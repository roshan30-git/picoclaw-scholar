package providers

import (
	"context"
	"testing"

	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
)

type mockProvider struct {
	lastModel string
	err       error
}

func (m *mockProvider) Chat(ctx context.Context, messages []tools.Message, toolsDef []tools.ToolDefinition, model string, options map[string]any) (*tools.LLMResponse, error) {
	m.lastModel = model
	return &tools.LLMResponse{}, m.err
}

func (m *mockProvider) GetDefaultModel() string {
	return "mock-default"
}

func TestNewModelRouter(t *testing.T) {
	mock := &mockProvider{}
	router := NewModelRouter(mock)
	if router == nil {
		t.Fatal("NewModelRouter returned nil")
	}
	if router.baseProvider != mock {
		t.Error("NewModelRouter did not set baseProvider correctly")
	}
}

func TestModelRouter_GetDefaultModel(t *testing.T) {
	router := NewModelRouter(&mockProvider{})
	got := router.GetDefaultModel()
	want := string(TierStandard)
	if got != want {
		t.Errorf("GetDefaultModel() = %q, want %q", got, want)
	}
}

func TestModelRouter_Chat(t *testing.T) {
	tests := []struct {
		name           string
		requestedModel string
		expectedModel  string
	}{
		{
			name:           "default to TierStandard",
			requestedModel: "",
			expectedModel:  string(TierStandard),
		},
		{
			name:           "use requested model",
			requestedModel: "custom-model",
			expectedModel:  "custom-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockProvider{}
			router := NewModelRouter(mock)
			_, _ = router.Chat(context.Background(), nil, nil, tt.requestedModel, nil)
			if mock.lastModel != tt.expectedModel {
				t.Errorf("Chat() sent model %q, want %q", mock.lastModel, tt.expectedModel)
			}
		})
	}
}

func TestModelRouter_RouteCost(t *testing.T) {
	tests := []struct {
		hint string
		want ModelTier
	}{
		{"indexer", TierFree},
		{"scheduler", TierFree},
		{"strategist", TierAdvanced},
		{"deep_analysis", TierAdvanced},
		{"random", TierStandard},
		{"", TierStandard},
	}

	router := NewModelRouter(&mockProvider{})
	for _, tt := range tests {
		t.Run(tt.hint, func(t *testing.T) {
			if got := router.RouteCost(tt.hint); got != string(tt.want) {
				t.Errorf("RouteCost(%q) = %q, want %q", tt.hint, got, string(tt.want))
			}
		})
	"testing"
)

func TestModelRouter_RouteCost(t *testing.T) {
	tests := []struct {
		taskHint string
		want     string
	}{
		{"indexer", string(TierFree)},
		{"scheduler", string(TierFree)},
		{"strategist", string(TierAdvanced)},
		{"deep_analysis", string(TierAdvanced)},
		{"", string(TierStandard)},
		{"unknown", string(TierStandard)},
		{"visual", string(TierStandard)},
	}

	router := &ModelRouter{}

	for _, tt := range tests {
		t.Run(tt.taskHint, func(t *testing.T) {
			if got := router.RouteCost(tt.taskHint); got != tt.want {
				t.Errorf("RouteCost(%q) = %v, want %v", tt.taskHint, got, tt.want)
			}
		})
	}
}

func TestModelRouter_GetDefaultModel(t *testing.T) {
	router := &ModelRouter{}
	want := string(TierStandard)
	if got := router.GetDefaultModel(); got != want {
		t.Errorf("GetDefaultModel() = %v, want %v", got, want)
	}
}
