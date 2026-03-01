package config

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
