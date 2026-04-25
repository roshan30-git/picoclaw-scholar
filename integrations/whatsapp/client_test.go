package whatsapp

import (
	"os"
	"testing"
)

func TestGetOwnerEnv(t *testing.T) {
	const envKey = "STUDYCLAW_OWNER_NUMBER"
	orig, exists := os.LookupEnv(envKey)
	t.Cleanup(func() {
		if exists {
			os.Setenv(envKey, orig)
		} else {
			os.Unsetenv(envKey)
		}
	})

	tests := []struct {
		name     string
		envVal   string
		expected string
	}{
		{"Environment variable set", "919876543210", "919876543210"},
		{"Environment variable empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(envKey, tt.envVal)
			got := GetOwnerEnv()
			if got != tt.expected {
				t.Errorf("GetOwnerEnv() = %q, want %q", got, tt.expected)
			}
		})
	}

	t.Run("Environment variable unset", func(t *testing.T) {
		os.Unsetenv(envKey)
		got := GetOwnerEnv()
		if got != "" {
			t.Errorf("GetOwnerEnv() = %q, want empty string", got)
		}
	})
}
