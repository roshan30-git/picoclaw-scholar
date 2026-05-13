package utils

import "testing"

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		limit    int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello", 4, "h..."},
		{"hello", 3, "..."},
		{"hello", 2, "he"},
		{"", 5, ""},
	}

	for _, tt := range tests {
		result := Truncate(tt.input, tt.limit)
		if result != tt.expected {
			t.Errorf("Truncate(%q, %d) = %q; want %q", tt.input, tt.limit, result, tt.expected)
		}
	}
}
