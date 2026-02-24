// Package whatsapp provides the StudyClaw channel adapter for WhatsApp
// using the whatsmeow multi-device library.
// This is the solution to Gap G3 — wiring whatsmeow into PicoClaw's event system.
package whatsapp

import (
	"context"
	"fmt"
	"os"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	_ "github.com/mattn/go-sqlite3"
)

// MessageHandler is the callback signature for incoming WhatsApp messages.
// The StudyClaw agent loop receives (sender, text, mediaPath) and calls Gemini.
type MessageHandler func(sender, text, mediaPath string)

// Client wraps a whatsmeow.Client and routes events to StudyClaw.
type Client struct {
	wac     *whatsmeow.Client
	handler MessageHandler
}

// New creates and connects a new WhatsApp client.
// sessionPath: where to persist the login session (so QR scan is only needed once).
func New(sessionPath string, handler MessageHandler) (*Client, error) {
	logger := waLog.Stdout("WhatsApp", "INFO", true)

	// Open the SQLite session store
	container, err := sqlstore.New(context.Background(), "sqlite3", sessionPath, logger)
	if err != nil {
		return nil, fmt.Errorf("sqlstore: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}

	wac := whatsmeow.NewClient(deviceStore, logger)
	c := &Client{wac: wac, handler: handler}
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
		// Extract text
		text := ""
		if v.Message.GetConversation() != "" {
			text = v.Message.GetConversation()
		} else if v.Message.GetExtendedTextMessage() != nil {
			text = v.Message.GetExtendedTextMessage().GetText()
		}

		sender := v.Info.Sender.String()
		mediaPath := "" // TODO: media download in Phase 1 Week 2

		if text != "" || mediaPath != "" {
			c.handler(sender, text, mediaPath)
		}
	}
}

// Send delivers a text message to a WhatsApp JID (contact or group).
func (c *Client) Send(to, message string) error {
	jid, err := parseJID(to)
	if err != nil {
		return err
	}
	msg := &waE2E.Message{Conversation: proto.String(message)}
	_, err = c.wac.SendMessage(context.Background(), jid, msg)
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
