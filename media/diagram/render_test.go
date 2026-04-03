package diagram

import (
	"strings"
	"testing"
)

func TestRenderForTelegram(t *testing.T) {
	mermaid := "graph TD; A-->B;"
	caption := "Test Diagram"
	webAppURL := "https://test.app/diagram"

	msg := RenderForTelegram(mermaid, caption, webAppURL)
	if !strings.Contains(msg.InlineKeyboard[0][0].WebApp.URL, webAppURL) {
		t.Errorf("Expected URL %s, got %s", webAppURL, msg.InlineKeyboard[0][0].WebApp.URL)
	}
	if !strings.Contains(msg.Text, caption) {
		t.Errorf("Expected caption %s in text, got %s", caption, msg.Text)
	}
}

func TestRenderForWhatsApp(t *testing.T) {
	mermaid := "graph TD; A-->B;"
	caption := "Test Diagram"

	res := RenderForWhatsApp(mermaid, caption)
	if !strings.Contains(res, mermaid) {
		t.Errorf("Expected mermaid syntax in response, got %s", res)
	}
	if !strings.Contains(res, caption) {
		t.Errorf("Expected caption %s in response, got %s", caption, res)
	}
}
