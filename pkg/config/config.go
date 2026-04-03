package config

import (
	"os"
	"strings"
)

const DefaultTelegramWebAppURL = "https://your-studyclaw-miniapp.vercel.app/diagram"

type StudentProfile struct {
	Name         string   `json:"name"`
	Semester     int      `json:"semester"`
	Subjects     []string `json:"subjects"`
	WeakTopics   []string `json:"weak_topics"`
	LearningPace string   `json:"learning_pace"`
}

type TelegramConfig struct {
	Token     string   `json:"token"`
	Enabled   bool     `json:"enabled"`
	Proxy     string   `json:"proxy"`
	AllowFrom []string `json:"allow_from"`
}

type WhatsAppConfig struct {
	Enabled bool `json:"enabled"`
}

type GenericChannelConfig struct {
	Enabled bool `json:"enabled"`
}

type ChannelsConfig struct {
	Telegram TelegramConfig       `json:"telegram"`
	WhatsApp WhatsAppConfig       `json:"whatsapp"`
	Feishu   GenericChannelConfig `json:"feishu"`
	Discord  GenericChannelConfig `json:"discord"`
	Slack    GenericChannelConfig `json:"slack"`
}

type AgentDefaults struct {
	Model    string `json:"model"`
	Provider string `json:"provider"`
}

type AgentsConfig struct {
	Defaults AgentDefaults `json:"defaults"`
}

type Config struct {
	ModelName         string         `json:"model_name"`
	MaxTokens         int            `json:"max_tokens"`
	SystemPrompt      string         `json:"system_prompt"`
	PromptDir         string         `json:"prompt_dir"`
	TelegramWebAppURL string         `json:"telegram_webapp_url"`
	StudentProfile    StudentProfile `json:"student_profile"`
	AllowedGroupJIDs  []string       `json:"allowed_group_jids"`
	PassiveGroupJIDs  []string       `json:"passive_group_jids"`
	Channels          ChannelsConfig `json:"channels"`
	Agents            AgentsConfig   `json:"agents"`
}

func DefaultConfig() *Config {
	return &Config{
		ModelName:         "gemini-2.0-flash",
		MaxTokens:         8192,
		SystemPrompt:      "",
		PromptDir:         "workspace/PROMPTS",
		TelegramWebAppURL: DefaultTelegramWebAppURL,
		StudentProfile: StudentProfile{
			Name:         "Student",
			LearningPace: "medium",
		},
		AllowedGroupJIDs: []string{},
		PassiveGroupJIDs: []string{},
		Channels: ChannelsConfig{
			Telegram: TelegramConfig{
				Enabled: true,
			},
		},
		Agents: AgentsConfig{
			Defaults: AgentDefaults{
				Model:    "gemini-2.0-flash",
				Provider: "gemini",
			},
		},
	}
}

// LoadConfig initializes the configuration from environment variables,
// falling back to defaults where necessary.
func LoadConfig() *Config {
	cfg := DefaultConfig()

	if model := os.Getenv("MODEL_NAME"); model != "" {
		cfg.ModelName = model
	}
	if promptDir := os.Getenv("PROMPT_DIR"); promptDir != "" {
		cfg.PromptDir = promptDir
	}
	if webAppURL := os.Getenv("TELEGRAM_WEBAPP_URL"); webAppURL != "" {
		cfg.TelegramWebAppURL = webAppURL
	}
	if ownerNumber := os.Getenv("STUDYCLAW_OWNER_NUMBER"); ownerNumber != "" {
		// Populate allowed JIDs if owner number is provided
		cfg.AllowedGroupJIDs = append(cfg.AllowedGroupJIDs, ownerNumber)
	}

	if allowedGroups := os.Getenv("STUDYCLAW_ALLOWED_GROUPS"); allowedGroups != "" {
		groups := strings.Split(allowedGroups, ",")
		for _, g := range groups {
			trimmed := strings.TrimSpace(g)
			if trimmed != "" {
				cfg.AllowedGroupJIDs = append(cfg.AllowedGroupJIDs, trimmed)
			}
		}
	}

	if passiveGroups := os.Getenv("STUDYCLAW_PASSIVE_GROUPS"); passiveGroups != "" {
		groups := strings.Split(passiveGroups, ",")
		for _, g := range groups {
			trimmed := strings.TrimSpace(g)
			if trimmed != "" {
				cfg.PassiveGroupJIDs = append(cfg.PassiveGroupJIDs, trimmed)
			}
		}
	}

	// Provider-specific config
	if provider := os.Getenv("LLM_PROVIDER"); provider != "" {
		cfg.Agents.Defaults.Provider = provider
	}

	return cfg
}
