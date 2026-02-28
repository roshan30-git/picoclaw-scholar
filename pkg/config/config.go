package config

const DefaultTelegramWebAppURL = "https://your-studyclaw-miniapp.vercel.app/diagram"

type Config struct {
	ModelName         string `json:"model_name"`
	MaxTokens         int    `json:"max_tokens"`
	SystemPrompt      string `json:"system_prompt"`
	PromptDir         string `json:"prompt_dir"`
	TelegramWebAppURL string `json:"telegram_webapp_url"`
}

func DefaultConfig() *Config {
	return &Config{
		ModelName:         "gemini-2.0-flash",
		MaxTokens:         8192,
		SystemPrompt:      "",
		PromptDir:         "workspace/PROMPTS",
		TelegramWebAppURL: DefaultTelegramWebAppURL,
	}
}
