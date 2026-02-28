package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Gemini    GeminiConfig    `json:"gemini"`
	WhatsApp  WhatsAppConfig  `json:"whatsapp"`
	Scheduler SchedulerConfig `json:"scheduler"`
}

type GeminiConfig struct {
	APIKey string `json:"api_key"`
}

type WhatsAppConfig struct {
	OwnerNumber string `json:"owner_number"`
}

type SchedulerConfig struct {
	DailyQuizTime string `json:"daily_quiz_time"`
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("user home dir: %w", err)
	}

	configPath := filepath.Join(home, ".studyclaw", "config.json")

	// If config doesn't exist, return a default empty config or error?
	// For now, let's return a default config if file not found, but log a warning.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	return &cfg, nil
}
