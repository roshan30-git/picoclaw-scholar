package providers

import (
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
