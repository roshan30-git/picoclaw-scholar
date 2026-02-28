package channels

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/roshan30-git/picoclaw-scholar/pkg/bus"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramChannel struct {
	BaseChannel
	bot *tgbotapi.BotAPI
}

func NewTelegramChannel(b *bus.MessageBus) (*TelegramChannel, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("telegram connect: %w", err)
	}

	log.Printf("Authorized on Telegram account %s", bot.Self.UserName)

	return &TelegramChannel{
		BaseChannel: BaseChannel{
			name: "telegram",
			bus:  b,
		},
		bot: bot,
	}, nil
}

func (t *TelegramChannel) Start(ctx context.Context) error {
	t.SetRunning(true)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	go func() {
		log.Println("Telegram channel listener started")
		for {
			select {
			case <-ctx.Done():
				t.SetRunning(false)
				return
			case update := <-updates:
				if update.Message != nil {
					chatID := fmt.Sprintf("%d", update.Message.Chat.ID)
					from := update.Message.From.UserName
					if from == "" {
						from = update.Message.From.FirstName
					}
					
					// Route to bus
					t.HandleMessage(from, chatID, update.Message.Text, nil, nil)
				}
			}
		}
	}()

	return nil
}

func (t *TelegramChannel) Stop(ctx context.Context) error {
	log.Println("Telegram channel stopping...")
	t.bot.StopReceivingUpdates()
	t.SetRunning(false)
	return nil
}

func (t *TelegramChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	if !t.IsRunning() {
		return fmt.Errorf("telegram channel is not running")
	}

	chatIDInt := int64(0)
	fmt.Sscanf(msg.ChatID, "%d", &chatIDInt)

	tgMsg := tgbotapi.NewMessage(chatIDInt, msg.Content)
	tgMsg.ParseMode = "Markdown"
	
	if msg.VisualID != "" {
		// Create a WebApp button linking to the diagram viewer
		// In production, this must be a valid HTTPS URL (e.g. via ngrok or proper domain)
		webAppURL := fmt.Sprintf("https://studyclaw.app/viewer?id=%s&type=%s", msg.VisualID, msg.VisualType)
		btn := tgbotapi.InlineKeyboardButton{
			Text:   "🔍 Tap to View Diagram/Formula",
			WebApp: &tgbotapi.WebAppInfo{URL: webAppURL},
		}
		tgMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(btn),
		)
	}

	_, err := t.bot.Send(tgMsg)
	return err
}

