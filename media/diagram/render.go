// Package diagram provides visual content delivery via Telegram Mini Apps.
// Strategy (replaces Mermaid→PNG): Telegram's built-in Mini App / WebApp API
// can render rich HTML+Mermaid diagrams inline — no image conversion needed.
// For WhatsApp fallback: plain Mermaid code block in text.
package diagram

import (
	"fmt"
	"strings"
)

// TelegramWebAppURL is the hosted StudyClaw Mini App that renders diagrams.
// Replace with your deployed Telegram Mini App URL.
const TelegramWebAppURL = "https://your-studyclaw-miniapp.vercel.app/diagram"

// RenderForTelegram returns an inline keyboard button that opens the Telegram
// Mini App (WebApp) to render the Mermaid diagram interactively.
// The diagram code is passed as a URL-encoded query param.
func RenderForTelegram(mermaidSyntax, caption string) TelegramMessage {
	encoded := strings.ReplaceAll(mermaidSyntax, "\n", "%0A")
	url := fmt.Sprintf("%s?code=%s", TelegramWebAppURL, encoded)

	return TelegramMessage{
		Text: fmt.Sprintf("📊 *%s*\n\nTap below to view the interactive diagram:", caption),
		InlineKeyboard: [][]InlineButton{
			{
				{
					Text:   "🔬 Open Diagram",
					WebApp: &WebAppInfo{URL: url},
				},
			},
		},
	}
}

// RenderForWhatsApp returns a plain text Mermaid code block as fallback.
// WhatsApp cannot render Mini Apps, so we send the raw syntax.
func RenderForWhatsApp(mermaidSyntax, caption string) string {
	return fmt.Sprintf("📊 *%s*\n```\n%s\n```\n_(Use Telegram for interactive diagram view)_", caption, mermaidSyntax)
}

// TelegramMessage represents the structured message payload for Telegram Bot API.
type TelegramMessage struct {
	Text           string
	InlineKeyboard [][]InlineButton
}

// InlineButton is a Telegram inline keyboard button.
type InlineButton struct {
	Text         string       `json:"text"`
	WebApp       *WebAppInfo  `json:"web_app,omitempty"`
	CallbackData string       `json:"callback_data,omitempty"`
}

// WebAppInfo holds the Telegram Mini App (WebApp) URL.
type WebAppInfo struct {
	URL string `json:"url"`
}
