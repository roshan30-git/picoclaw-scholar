package study

import (
	"fmt"
	"strings"

	"github.com/roshan30-git/picoclaw-scholar/pkg/database"
)

// TopicScore represents a topic's mastery level.
type TopicScore struct {
	Topic    string
	AvgScore float64
	Attempts int
}

// FormatWeakTopicHeatmap queries learning_profile and renders an emoji grid.
// 🔴 score < 40% | 🟡 40-70% | 🟢 > 70%
func FormatWeakTopicHeatmap(db *database.DB) string {
	rows, err := db.Conn().Query(
		`SELECT topic, avg_score, attempts FROM learning_profile ORDER BY avg_score ASC`,
	)
	if err != nil || rows == nil {
		return "📊 No quiz data yet! Send me a PDF and try `quiz me` first."
	}
	defer rows.Close()

	var topics []TopicScore
	for rows.Next() {
		var t TopicScore
		if err := rows.Scan(&t.Topic, &t.AvgScore, &t.Attempts); err != nil {
			continue
		}
		topics = append(topics, t)
	}

	if len(topics) == 0 {
		return "📊 No quiz data yet! Send me a PDF and try `quiz me` to build your profile."
	}

	var b strings.Builder
	b.WriteString("📊 *Weak Topic Heatmap*\n\n")

	for _, t := range topics {
		pct := t.AvgScore * 100
		var emoji string
		switch {
		case pct < 40:
			emoji = "🔴"
		case pct < 70:
			emoji = "🟡"
		default:
			emoji = "🟢"
		}

		// ASCII progress bar
		filled := int(pct / 10)
		if filled < 0 {
			filled = 0
		}
		if filled > 10 {
			filled = 10
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", 10-filled)

		b.WriteString(fmt.Sprintf("%s %s [%s] %.0f%% (%d quizzes)\n",
			emoji, t.Topic, bar, pct, t.Attempts))
	}

	return b.String()
}

// FormatProgressBars queries learning_profile and renders ASCII progress bars per subject.
func FormatProgressBars(db *database.DB) string {
	rows, err := db.Conn().Query(
		`SELECT topic, avg_score, attempts FROM learning_profile ORDER BY topic ASC`,
	)
	if err != nil || rows == nil {
		return "📈 No progress data yet! Take some quizzes first."
	}
	defer rows.Close()

	var topics []TopicScore
	for rows.Next() {
		var t TopicScore
		if err := rows.Scan(&t.Topic, &t.AvgScore, &t.Attempts); err != nil {
			continue
		}
		topics = append(topics, t)
	}

	if len(topics) == 0 {
		return "📈 No progress data yet! Take some quizzes to see your progress."
	}

	var b strings.Builder
	b.WriteString("📈 *Subject Progress*\n\n")

	for _, t := range topics {
		pct := t.AvgScore * 100
		filled := int(pct / 5)
		if filled < 0 {
			filled = 0
		}
		if filled > 20 {
			filled = 20
		}
		bar := strings.Repeat("████", filled/4) + strings.Repeat("░░░░", (20-filled)/4)
		b.WriteString(fmt.Sprintf("%-20s [%s] %3.0f%%\n", t.Topic, bar, pct))
	}

	return b.String()
}
