package telegram

import (
	"testing"
)

func TestMarkdownToTelegramHTML(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected string
	}{
		{
			name:     "Headers",
			markdown: "# Header 1\n## Header 2",
			expected: "Header 1\nHeader 2",
		},
		{
			name:     "Bold and Italic",
			markdown: "This is **bold** and _italic_.",
			expected: "This is <b>bold</b> and <i>italic</i>.",
		},
		{
			name:     "Links",
			markdown: "[Google](https://google.com)",
			expected: `<a href="https://google.com">Google</a>`,
		},
		{
			name:     "Inline Code",
			markdown: "Use `go test` to run tests.",
			expected: "Use <code>go test</code> to run tests.",
		},
		{
			name:     "Code Blocks",
			markdown: "```go\nfmt.Println(\"Hello\")\n```",
			expected: "<pre><code>fmt.Println(\"Hello\")\n</code></pre>",
		},
		{
			name:     "Strikethrough and Lists",
			markdown: "~~deleted~~\n- item 1\n* item 2",
			expected: "<s>deleted</s>\n• item 1\n• item 2",
		},
		{
			name:     "HTML Escaping",
			markdown: "3 < 5 & 4 > 2",
			expected: "3 &lt; 5 &amp; 4 &gt; 2",
		},
		{
			name:     "Mixed Formatting",
			markdown: "**Bold** and `code` in the same line.",
			expected: "<b>Bold</b> and <code>code</code> in the same line.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := markdownToTelegramHTML(tt.markdown)
			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}
