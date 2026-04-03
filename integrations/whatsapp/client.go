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
	passiveGroups []string
	ocr           *study.OCRPipeline
}

// New creates and connects a new WhatsApp client.
// Uses phone-number pairing code so no QR scan is needed — perfect for Termux.
func New(ctx context.Context, sessionPath string, msgBus *bus.MessageBus, allowedGroups []string, passiveGroups []string, ocrPipeline *study.OCRPipeline) (*Client, error) {
	logger := waLog.Stdout("WhatsApp", "INFO", true)

	// Append modernc SQLite pragmas: Enable WAL to prevent locking errors during history sync
	connectionString := sessionPath + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	container, err := sqlstore.New(context.Background(), "sqlite", connectionString, logger)
	if err != nil {
		return nil, fmt.Errorf("sqlstore: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}

	wac := whatsmeow.NewClient(deviceStore, logger)
	c := &Client{wac: wac, bus: msgBus, allowedGroups: allowedGroups, passiveGroups: passiveGroups, ocr: ocrPipeline}
	wac.AddEventHandler(c.handleEvent)

	if wac.Store.ID == nil {
		// ── NOT LOGGED IN ─────────────────────────────────────────
		// Use pairing code instead of QR — works perfectly in Termux.
		// The user types their number, gets a code, enters it in WhatsApp.
		if err := wac.Connect(); err != nil {
			return nil, fmt.Errorf("connect: %w", err)
		}

		phoneNumber := os.Getenv("STUDYCLAW_OWNER_NUMBER")
		if phoneNumber == "" {
			fmt.Print("\n📱 Enter your WhatsApp phone number (with country code, e.g. 919876543210): ")
			fmt.Scanln(&phoneNumber)
		}

		code, err := wac.PairPhone(context.Background(), phoneNumber, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
		if err != nil {
			return nil, fmt.Errorf("pair phone: %w", err)
		}

		fmt.Printf("\n╔══════════════════════════════════════╗\n")
		fmt.Printf("║  WhatsApp Pairing Code: %s  ║\n", code)
		fmt.Printf("╚══════════════════════════════════════╝\n\n")
		fmt.Println("👆 Go to WhatsApp → Linked Devices → Link a Device → Link with phone number")
		fmt.Println("   Enter the code above. You have 60 seconds.")
		fmt.Println()

		// Wait for pairing to complete
		linked := make(chan struct{}, 1)
		wac.AddEventHandler(func(evt interface{}) {
			if _, ok := evt.(*events.LoggedOut); ok {
				fmt.Println("❌ Pairing failed or was canceled.")
				select {
				case linked <- struct{}{}:
				default:
				}
			}
			if wac.Store.ID != nil {
				select {
				case linked <- struct{}{}:
				default:
				}
			}
		})

		// Also set a timeout for pairing safety
		select {
		case <-linked:
			if wac.Store.ID != nil {
				fmt.Println("✅ Successfully linked WhatsApp!")
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	} else {
		// ── ALREADY LOGGED IN ─────────────────────────────────────
		if err := wac.Connect(); err != nil {
			return nil, fmt.Errorf("reconnect: %w", err)
		}
		go c.listGroups()
	}

	return c, nil
}

func (c *Client) listGroups() {
	groups, err := c.wac.GetJoinedGroups(context.Background())
	if err != nil {
		fmt.Printf("Warning: Failed to list joined groups: %v\n", err)
		return
	}

	if len(groups) > 0 {
		fmt.Println("\n📋 Joined WhatsApp Groups (for configuration):")
		for _, g := range groups {
			fmt.Printf("   - %s (JID: %s)\n", g.Name, g.JID)
		}
		fmt.Println("")
	}
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
		if v.Info.IsGroup && !c.isAllowedGroup(chatID) {
			return // Ignore non-allowed group message
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
			os.MkdirAll(dir, 0700)

			// Quick unique filename
			filename := fmt.Sprintf("%s_%s%s", sender, v.Info.ID, fileExt)
			mediaPath = dir + "/" + filename
			os.WriteFile(mediaPath, mediaData, 0600)
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

func (c *Client) isAllowedGroup(chatID string) bool {
	if len(c.allowedGroups) == 0 {
		return true
	}
	for _, g := range c.allowedGroups {
		if chatID == g || chatID == g+"@g.us" {
			return true
		}
	}
	return false
}

func (c *Client) isPassiveGroup(chatID string) bool {
	for _, g := range c.passiveGroups {
		if chatID == g || chatID == g+"@g.us" {
			return true
		}
	}
	return false
}

var diagramRegex = regexp.MustCompile("(?s)```mermaid(.*?)```")

// Send delivers a text message to a WhatsApp JID (contact or group).
func (c *Client) Send(ctx context.Context, outMsg bus.OutboundMessage) error {
	isPassive := c.isPassiveGroup(outMsg.ChatID)

	jid, err := parseJID(outMsg.ChatID)
	if err != nil {
		return err
	}

	targetJID := jid
	content := outMsg.Content

	if isPassive {
		// Redirect to owner
		owner := GetOwnerEnv()
		if owner == "" {
			fmt.Printf("[PASSIVE BLOCK] Response suppressed for group %s: %s\n", outMsg.ChatID, content)
			return nil
		}
		targetJID, _ = parseJID(owner)
		content = fmt.Sprintf("⚠️ [PASSIVE MODE RESPONSE FOR %s]\n\n%s", outMsg.ChatID, content)
	}

	if outMsg.VisualID != "" {
		content += fmt.Sprintf("\n\n🔍 View Diagram/Formula here: https://studyclaw.app/viewer?id=%s&type=%s", outMsg.VisualID, outMsg.VisualType)
	}

	waMsg := &waE2E.Message{Conversation: proto.String(content)}
	_, err = c.wac.SendMessage(context.Background(), targetJID, waMsg)
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
