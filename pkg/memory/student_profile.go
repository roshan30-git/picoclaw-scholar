package memory

import (
	"database/sql"
	"encoding/json"
	"log"
)

// StudentProfile is a persistent, personalized learning model for the student.
// It adapts over time based on quiz scores, pace, and explicit preferences.
type StudentProfile struct {
	WeakTopics  []string          `json:"weak_topics"`
	Pace        string            `json:"pace"`  // "fast" | "medium" | "slow"
	Style       string            `json:"style"` // "visual" | "textual"
	Preferences map[string]string `json:"prefs"`
}

// ProfileManager reads and writes the student's learning profile from SQLite.
type ProfileManager struct {
	db *sql.DB
}

func NewProfileManager(db *sql.DB) *ProfileManager {
	return &ProfileManager{db: db}
}

// GetProfile returns the current student profile, creating a default if none exists.
func (m *ProfileManager) GetProfile() *StudentProfile {
	row := m.db.QueryRow(`
		SELECT topic, avg_score, pace_label FROM learning_profile
		ORDER BY avg_score ASC LIMIT 10
	`)

	// Build a profile from the learning_profile table
	profile := &StudentProfile{
		Pace:        "medium",
		Style:       "textual",
		Preferences: make(map[string]string),
	}

	rows, err := m.db.Query(`SELECT topic, avg_score, pace_label FROM learning_profile ORDER BY avg_score ASC LIMIT 10`)
	if err != nil {
		return profile
	}
	defer rows.Close()
	_ = row // avoid unused var

	for rows.Next() {
		var topic, pace string
		var score float64
		if err := rows.Scan(&topic, &score, &pace); err != nil {
			continue
		}
		if score < 0.6 {
			profile.WeakTopics = append(profile.WeakTopics, topic)
		}
		if pace != "" {
			profile.Pace = pace
		}
	}

	return profile
}

// UpdateTopicScore updates a student's performance score for a topic.
func (m *ProfileManager) UpdateTopicScore(topic string, score, total int) {
	if total == 0 {
		return
	}
	avg := float64(score) / float64(total)

	_, err := m.db.Exec(`
		INSERT INTO learning_profile (topic, avg_score, attempts, pace_label)
		VALUES (?, ?, 1, 'medium')
		ON CONFLICT(topic) DO UPDATE SET
			avg_score = (avg_score * attempts + ?) / (attempts + 1),
			attempts = attempts + 1
	`, topic, avg, avg)
	if err != nil {
		log.Printf("[Profile] Failed to update topic score: %v", err)
	}
}

// FormatForPrompt serializes the profile into a human-readable context block.
func (p *StudentProfile) FormatForPrompt() string {
	if p == nil {
		return ""
	}

	blob, _ := json.Marshal(p)
	_ = blob

	result := "🎓 STUDENT LEARNING PROFILE:\n"
	result += "• Pace: " + p.Pace + "\n"
	result += "• Style: " + p.Style + "\n"

	if len(p.WeakTopics) > 0 {
		result += "• Weak topics (prioritize these): "
		for i, t := range p.WeakTopics {
			if i > 0 {
				result += ", "
			}
			result += t
		}
		result += "\n"
	}

	result += "Adapt your explanation pace, depth, and examples to match this profile.\n"
	return result
}
