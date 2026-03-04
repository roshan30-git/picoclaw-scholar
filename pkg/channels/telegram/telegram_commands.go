package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/mymmrac/telego"

	"github.com/roshan30-git/picoclaw-scholar/pkg/config"
	pkgdb "github.com/roshan30-git/picoclaw-scholar/pkg/database"
)

type TelegramCommander interface {
	Help(ctx context.Context, message telego.Message) error
	Start(ctx context.Context, message telego.Message) error
	Show(ctx context.Context, message telego.Message) error
	List(ctx context.Context, message telego.Message) error
}

type cmd struct {
	bot    *telego.Bot
	config *config.Config
	db     *pkgdb.DB
}

func NewTelegramCommands(bot *telego.Bot, cfg *config.Config, db *pkgdb.DB) TelegramCommander {
	return &cmd{
		bot:    bot,
		config: cfg,
		db:     db,
	}
}

func commandArgs(text string) string {
	parts := strings.SplitN(text, " ", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func (c *cmd) Help(ctx context.Context, message telego.Message) error {
	msg := `/start - Start the bot
/help - Show this help message
/show [model|channel] - Show current configuration
/list [models|channels] - List available options
	`
	_, err := c.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: message.Chat.ID},
		Text:   msg,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	})
	return err
}

func (c *cmd) Start(ctx context.Context, message telego.Message) error {
	chatIDStr := fmt.Sprintf("%d", message.Chat.ID)

	// Check if user is already onboarded
	profile, err := c.db.GetUserProfile(chatIDStr)

	if err == nil && profile.OnboardingComplete {
		// ── RETURNING USER ──────────────────────────────────────────
		keyboard := tu.InlineKeyboard(
			tu.InlineKeyboardRow(
				tu.InlineKeyboardButton("🎯 Start Quiz").WithCallbackData("action_quiz"),
				tu.InlineKeyboardButton("📅 My Deadlines").WithCallbackData("action_deadlines"),
			),
			tu.InlineKeyboardRow(
				tu.InlineKeyboardButton("📊 Weekly Report").WithCallbackData("action_report"),
				tu.InlineKeyboardButton("🔍 Search Notes").WithCallbackData("action_search"),
			),
			tu.InlineKeyboardRow(
				tu.InlineKeyboardButton("📥 Upload PDF/Image").WithCallbackData("action_upload"),
				tu.InlineKeyboardButton("❓ Help").WithCallbackData("action_help"),
			),
		)
		text := fmt.Sprintf("👋 Welcome back, <b>%s</b>! (%s · %s)\n\n<i>What would you like to do today?</i>",
			message.From.FirstName, profile.Semester, profile.University)

		_, err = c.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID:    telego.ChatID{ID: message.Chat.ID},
			Text:      text,
			ParseMode: telego.ModeHTML,
			ReplyMarkup: keyboard,
		})
		return err
	}

	// ── NEW USER ONBOARDING ──────────────────────────────────────────────
	if err == nil && profile.University != "" && profile.Semester == "" {
		text := "🎓 Almost done! <b>Which semester are you in?</b>"
		keyboard := tu.InlineKeyboard(
			tu.InlineKeyboardRow(
				tu.InlineKeyboardButton("Sem 1").WithCallbackData("sem_1st Semester"),
				tu.InlineKeyboardButton("Sem 2").WithCallbackData("sem_2nd Semester"),
				tu.InlineKeyboardButton("Sem 3").WithCallbackData("sem_3rd Semester"),
			),
			tu.InlineKeyboardRow(
				tu.InlineKeyboardButton("Sem 4").WithCallbackData("sem_4th Semester"),
				tu.InlineKeyboardButton("Sem 5").WithCallbackData("sem_5th Semester"),
				tu.InlineKeyboardButton("Sem 6").WithCallbackData("sem_6th Semester"),
			),
			tu.InlineKeyboardRow(
				tu.InlineKeyboardButton("Sem 7").WithCallbackData("sem_7th Semester"),
				tu.InlineKeyboardButton("Sem 8").WithCallbackData("sem_8th Semester"),
			),
		)
		_, err = c.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID}, Text: text, ParseMode: telego.ModeHTML, ReplyMarkup: keyboard,
		})
		return err
	}

	// Fresh start
	_ = c.db.SaveUserProfile(&pkgdb.UserProfile{ChatID: chatIDStr})

	welcomeText := `🦞 <b>Welcome to StudyClaw!</b>

Your autonomous AI study agent, running 24/7.

<b>Here's what I can do:</b>
🎯 <b>Quiz you</b> from your own college group notes
📥 <b>Index PDFs &amp; images</b> — just send them to me
🔍 <b>Search your notes</b> with smart semantic search
📐 <b>Draw diagrams</b> — flowcharts, circuits, ERDs
📅 <b>Track deadlines</b> and send you daily reminders
📊 <b>Weekly performance report</b> every Sunday
🌐 <b>Search the internet</b> for current events
📖 <b>Summarize group messages</b> from your college groups silently

——

To get started, just tell me: <b>What University or College do you attend?</b>`

	_, err = c.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: message.Chat.ID},
		Text:      welcomeText,
		ParseMode: telego.ModeHTML,
		ReplyMarkup: tu.InlineKeyboard(
			tu.InlineKeyboardRow(
				tu.InlineKeyboardButton("📌 Skip Setup (Just Chat)").WithCallbackData("action_skip_onboarding"),
			),
		),
	})
	return err
}

func (c *cmd) Show(ctx context.Context, message telego.Message) error {
	args := commandArgs(message.Text)
	if args == "" {
		_, err := c.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "Usage: /show [model|channel]",
			ReplyParameters: &telego.ReplyParameters{
				MessageID: message.MessageID,
			},
		})
		return err
	}

	var response string
	switch args {
	case "model":
		response = fmt.Sprintf("Current Model: %s (Provider: %s)",
			c.config.Agents.Defaults.Model,
			c.config.Agents.Defaults.Provider)
	case "channel":
		response = "Current Channel: telegram"
	default:
		response = fmt.Sprintf("Unknown parameter: %s. Try 'model' or 'channel'.", args)
	}

	_, err := c.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: message.Chat.ID},
		Text:   response,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	})
	return err
}

func (c *cmd) List(ctx context.Context, message telego.Message) error {
	args := commandArgs(message.Text)
	if args == "" {
		_, err := c.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "Usage: /list [models|channels]",
			ReplyParameters: &telego.ReplyParameters{
				MessageID: message.MessageID,
			},
		})
		return err
	}

	var response string
	switch args {
	case "models":
		provider := c.config.Agents.Defaults.Provider
		if provider == "" {
			provider = "configured default"
		}
		response = fmt.Sprintf("Configured Model: %s\nProvider: %s\n\nTo change models, update config.yaml",
			c.config.Agents.Defaults.Model, provider)

	case "channels":
		var enabled []string
		if c.config.Channels.Telegram.Enabled {
			enabled = append(enabled, "telegram")
		}
		if c.config.Channels.WhatsApp.Enabled {
			enabled = append(enabled, "whatsapp")
		}
		if c.config.Channels.Feishu.Enabled {
			enabled = append(enabled, "feishu")
		}
		if c.config.Channels.Discord.Enabled {
			enabled = append(enabled, "discord")
		}
		if c.config.Channels.Slack.Enabled {
			enabled = append(enabled, "slack")
		}
		response = fmt.Sprintf("Enabled Channels:\n- %s", strings.Join(enabled, "\n- "))

	default:
		response = fmt.Sprintf("Unknown parameter: %s. Try 'models' or 'channels'.", args)
	}

	_, err := c.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: message.Chat.ID},
		Text:   response,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	})
	return err
}
