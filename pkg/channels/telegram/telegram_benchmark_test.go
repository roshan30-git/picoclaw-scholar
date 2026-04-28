package telegram

import (
	"testing"
)

func BenchmarkExtractInlineCodes(b *testing.B) {
	text := "Here is some `inline code` and another `piece of code` to extract."
	for i := 0; i < b.N; i++ {
		extractInlineCodes(text)
	}
}

func BenchmarkMarkdownToTelegramHTML(b *testing.B) {
	text := `
# Header
This is a **bold** text and _italic_ text.
Check out this [link](https://example.com).
` + "```go\nfunc main() {}\n```" + `
And some ` + "`inline code`" + `.
- List item 1
- List item 2
`
	for i := 0; i < b.N; i++ {
		markdownToTelegramHTML(text)
	}
}
