package telegram

import (
	"testing"
)

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"no special chars", "hello world", "hello world"},
		{"ampersand", "fish & chips", "fish &amp; chips"},
		{"less than", "a < b", "a &lt; b"},
		{"greater than", "x > y", "x &gt; y"},
		{"mixed", "<b> & </b>", "&lt;b&gt; &amp; &lt;/b&gt;"},
		{"multiple", "&&<<>>", "&amp;&amp;&lt;&lt;&gt;&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeHTML(tt.input); got != tt.expected {
				t.Errorf("escapeHTML(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMarkdownToTelegramHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain text", "hello", "hello"},
		{"bold double star", "**bold**", "<b>bold</b>"},
		{"bold double underscore", "__bold__", "<b>bold</b>"},
		{"italic", "_italic_", "<i>italic</i>"},
		{"strikethrough", "~~strike~~", "<s>strike</s>"},
		{"link", "[google](https://google.com)", "<a href=\"https://google.com\">google</a>"},
		{"inline code", "`code`", "<code>code</code>"},
		{"inline code escaping", "`a < b`", "<code>a &lt; b</code>"},
		{"code block", "```\nfunc main() {}\n```", "<pre><code>func main() {}\n</code></pre>"},
		{"code block escaping", "```\n<div>\n```", "<pre><code>&lt;div&gt;\n</code></pre>"},
		{"mixed with escaping", "Check this: `a < b` & [link](url)", "Check this: <code>a &lt; b</code> &amp; <a href=\"url\">link</a>"},
		{"header removal", "### Header", "Header"},
		{"blockquote removal", "> Quote", "Quote"},
		{"list bullet", "- item", "• item"},
		{"list bullet star", "* item", "• item"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := markdownToTelegramHTML(tt.input); got != tt.expected {
				t.Errorf("markdownToTelegramHTML(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
