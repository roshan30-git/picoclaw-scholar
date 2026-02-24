package config

type Config struct {
	ModelName    string
	MaxTokens    int
	SystemPrompt string
	PromptDir    string
}

func DefaultConfig() *Config {
	return &Config{
		ModelName:    "gemini-2.0-flash",
		MaxTokens:    8192,
		SystemPrompt: "",
		PromptDir:    "workspace/PROMPTS",
	}
}
