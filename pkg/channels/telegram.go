package channels

import (
	"context"
	"log"

	"github.com/roshan30-git/picoclaw-scholar/pkg/bus"
)

type TelegramChannel struct {
	BaseChannel
}

func NewTelegramChannel(bus *bus.MessageBus) *TelegramChannel {
	return &TelegramChannel{
		BaseChannel: BaseChannel{
			name: "telegram",
			bus:  bus,
		},
	}
}

func (t *TelegramChannel) Start(ctx context.Context) error {
	log.Println("Telegram channel starting (Placeholder)...")
	t.SetRunning(true)
	return nil
}

func (t *TelegramChannel) Stop(ctx context.Context) error {
	log.Println("Telegram channel stopping...")
	t.SetRunning(false)
	return nil
}

func (t *TelegramChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	log.Printf("[Telegram Send] To %s", msg.ChatID)
	return nil
}
