package config

const DefaultTelegramWebAppURL = "https://your-studyclaw-miniapp.vercel.app/diagram"

type StudentProfile struct {
	Name         string   `json:"name"`
	Semester     int      `json:"semester"`
	Subjects     []string `json:"subjects"`
	WeakTopics   []string `json:"weak_topics"`
	LearningPace string   `json:"learning_pace"`
}

type Config struct {
	ModelName         string         `json:"model_name"`
	MaxTokens         int            `json:"max_tokens"`
	SystemPrompt      string         `json:"system_prompt"`
	PromptDir         string         `json:"prompt_dir"`
	TelegramWebAppURL string         `json:"telegram_webapp_url"`
	StudentProfile    StudentProfile `json:"student_profile"`
	AllowedGroupJIDs  []string       `json:"allowed_group_jids"`
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
	}
}
