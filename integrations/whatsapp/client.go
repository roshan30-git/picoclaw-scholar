// Package whatsapp provides the StudyClaw channel adapter for WhatsApp
// using the whatsmeow multi-device library.
// This is the solution to Gap G3 — wiring whatsmeow into PicoClaw's event system.
package whatsapp

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/roshan30-git/picoclaw-scholar/pkg/bus"
	"github.com/roshan30-git/picoclaw-scholar/pkg/study"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	_ "modernc.org/sqlite"
)

// Client wraps a whatsmeow.Client and routes events to StudyClaw via MassageBus.
type Client struct {
	wac           *whatsmeow.Client
	bus           *bus.MessageBus
	allowedGroups []string
	ocr           *study.OCRPipeline
}

// New creates and connects a new WhatsApp client.
// sessionPath: where to persist the login session (so QR scan is only needed once).
func New(sessionPath string, msgBus *bus.MessageBus, allowedGroups []string, ocrPipeline *study.OCRPipeline) (*Client, error) {
	logger := waLog.Stdout("WhatsApp", "INFO", true)

	// Open the SQLite session store — foreign keys must be enabled for whatsmeow schema migrations
	container, err := sqlstore.New(context.Background(), "sqlite", sessionPath+"?_pragma=foreign_keys(1)", logger)
	if err != nil {
		return nil, fmt.Errorf("sqlstore: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}

	wac := whatsmeow.NewClient(deviceStore, logger)
	c := &Client{wac: wac, bus: msgBus, allowedGroups: allowedGroups, ocr: ocrPipeline}
	wac.AddEventHandler(c.handleEvent)

	// If not logged in, show QR code in terminal (user scans once)
	if wac.Store.ID == nil {
		qrChan, _ := wac.GetQRChannel(context.Background())
		if err := wac.Connect(); err != nil {
			return nil, fmt.Errorf("connect: %w", err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Println("📱 Scan this QR code in WhatsApp (Linked Devices):")
				fmt.Println(evt.Code) // In real impl: render to terminal QR
			}
		}
	} else {
		if err := wac.Connect(); err != nil {
			return nil, fmt.Errorf("reconnect: %w", err)
		}
	}

	return c, nil
}

// handleEvent is the whatsmeow event listener.
// It only processes plain text and media messages from allowed senders.
func (c *Client) handleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		chatID := v.Info.Chat.String()
		sender := v.Info.Sender.String()

		// Debug log for JID discovery
		if os.Getenv("DEBUG") == "true" {
			fmt.Printf("[WHATSAPP DEBUG] Message from %s in chat %s\n", sender, chatID)
		}

		// Extract text
		text := ""
		if v.Message.GetConversation() != "" {
			text = v.Message.GetConversation()
		} else if v.Message.GetExtendedTextMessage() != nil {
			text = v.Message.GetExtendedTextMessage().GetText()
		}

		// Administrative Commands (Owner Only)
		isOwner := sender == GetOwnerEnv() || sender == GetOwnerEnv()+"@s.whatsapp.net"
		if isOwner && strings.TrimSpace(text) == "!getid" {
			c.Send(context.Background(), bus.OutboundMessage{
				ChatID:  chatID,
				Content: fmt.Sprintf("📍 Chat ID: %s", chatID),
			})
			return
		}

		// Group filtering logic
		if v.Info.IsGroup && len(c.allowedGroups) > 0 {
			allowed := false
			for _, allowedJID := range c.allowedGroups {
				if chatID == allowedJID || chatID == allowedJID+"@g.us" {
					allowed = true
					break
				}
			}
			if !allowed {
				return // Ignore non-allowed group message
			}
		}

		mediaPath := ""

		// Auto-save media
		var mediaData []byte
		var err error
		fileExt := ".bin"

		if v.Message.GetDocumentMessage() != nil {
			mediaData, err = c.wac.Download(context.Background(), v.Message.GetDocumentMessage())
			fileExt = ".pdf"
		} else if v.Message.GetImageMessage() != nil {
			mediaData, err = c.wac.Download(context.Background(), v.Message.GetImageMessage())
			fileExt = ".jpg"
		}

		if err == nil && len(mediaData) > 0 {
			home, _ := os.UserHomeDir()
			dir := home + "/.studyclaw/media"
			os.MkdirAll(dir, 0755)

			// Quick unique filename
			filename := fmt.Sprintf("%s_%s%s", sender, v.Info.ID, fileExt)
			mediaPath = dir + "/" + filename
			os.WriteFile(mediaPath, mediaData, 0644)
			text = fmt.Sprintf("[Media Saved: %s] %s", mediaPath, text)

			// Process OCR if image and pipeline exists
			if fileExt == ".jpg" && c.ocr != nil {
				extracted, err := c.ocr.ExtractAndSave(context.Background(), mediaPath)
				if err == nil && extracted != "" {
					text += "\n[Extracted OCR Text]:\n" + extracted
				}
			}
		}

		if text != "" || mediaPath != "" {
			c.bus.Publish(bus.InboundMessage{
				From:    sender,
				ChatID:  chatID,
				Content: text,
				Channel: "whatsapp",
			})
		}
	}
}

func (c *Client) Name() string { return "whatsapp" }

func (c *Client) Start(ctx context.Context) error {
	if !c.wac.IsConnected() {
		return c.wac.Connect()
	}
	return nil
}

func (c *Client) Stop(ctx context.Context) error {
	c.Disconnect()
	return nil
}

func (c *Client) IsRunning() bool {
	return c.wac.IsConnected()
}

var diagramRegex = regexp.MustCompile("(?s)```mermaid(.*?)```")

// Send delivers a text message to a WhatsApp JID (contact or group).
func (c *Client) Send(ctx context.Context, outMsg bus.OutboundMessage) error {
	jid, err := parseJID(outMsg.ChatID)
	if err != nil {
		return err
	}

	content := outMsg.Content
	if outMsg.VisualID != "" {
		content += fmt.Sprintf("\n\n🔍 View Diagram/Formula here: https://studyclaw.app/viewer?id=%s&type=%s", outMsg.VisualID, outMsg.VisualType)
	}

	waMsg := &waE2E.Message{Conversation: proto.String(content)}
	_, err = c.wac.SendMessage(context.Background(), jid, waMsg)
	return err
}

// Disconnect cleanly closes the WhatsApp connection.
func (c *Client) Disconnect() {
	c.wac.Disconnect()
}

func parseJID(s string) (types.JID, error) {
	// Simple parser — extend if needed for group JIDs
	return types.ParseJID(s)
}

// GetOwnerEnv reads the owner phone number from environment or config.
func GetOwnerEnv() string {
	return os.Getenv("STUDYCLAW_OWNER_NUMBER")
}
