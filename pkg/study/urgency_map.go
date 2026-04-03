package study

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// FormatUrgencyMap renders upcoming deadlines as a color-coded urgency list.
// 🔴 < 1 day | 🟡 1-3 days | 🟢 > 3 days
func FormatUrgencyMap(deadlines []Deadline) string {
	if len(deadlines) == 0 {
		return "📅 No upcoming deadlines! You're all clear. 🎉"
	}

	var b strings.Builder
	b.WriteString("📋 *Urgency Map*\n\n")

	for _, d := range deadlines {
		hours := time.Until(d.DueDate).Hours()
		days := hours / 24.0

		var emoji, label string
		switch {
		case hours <= 0:
			emoji = "⚫"
			label = "OVERDUE"
		case days < 1:
			emoji = "🔴"
			label = fmt.Sprintf("%dh left", int(math.Max(hours, 1)))
		case days < 3:
			emoji = "🟡"
			label = fmt.Sprintf("%.0fd left", days)
		default:
			emoji = "🟢"
			label = fmt.Sprintf("%.0fd left", days)
		}

		b.WriteString(fmt.Sprintf("%s %s — _%s_\n", emoji, d.Title, label))
	}

	return b.String()
}
