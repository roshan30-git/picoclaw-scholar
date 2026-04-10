package study

import (
	"context"
	"fmt"
	"time"

	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
)

// AddDeadlineTool allows the LLM to add a new exam or assignment to the student's calendar.
type AddDeadlineTool struct {
	tracker *DeadlineTracker
}

func NewAddDeadlineTool(tracker *DeadlineTracker) *AddDeadlineTool {
	return &AddDeadlineTool{tracker: tracker}
}

func (t *AddDeadlineTool) Name() string { return "add_deadline" }
func (t *AddDeadlineTool) Description() string {
	return "Add a new assignment, exam, or deadline to the student's tracker."
}

func (t *AddDeadlineTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title": map[string]any{
				"type":        "string",
				"description": "Title of the exam or assignment",
			},
			"days_from_now": map[string]any{
				"type":        "integer",
				"description": "Number of days from today when the deadline is due",
			},
		},
		"required": []string{"title", "days_from_now"},
	}
}

func (t *AddDeadlineTool) Execute(ctx context.Context, params map[string]any) *tools.ToolResult {
	title, ok := params["title"].(string)
	if !ok || title == "" {
		return tools.ErrorResult("title parameter is required")
	}

	days, ok := params["days_from_now"].(float64)
	if !ok {
		return tools.ErrorResult("days_from_now parameter is required")
	}

	dueDate := time.Now().Add(time.Duration(days) * 24 * time.Hour)
	err := t.tracker.AddDeadline(title, dueDate, "ai_agent")
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to add deadline: %v", err))
	}

	dateStr := dueDate.Format("Jan 02, 2006")
	return tools.SuccessResult(
		fmt.Sprintf("✅ Added deadline: '%s' due on %s.", title, dateStr),
		fmt.Sprintf("Added deadline: %s", title),
	)
}

// ViewDeadlinesTool allows the LLM to check what exams are coming up.
type ViewDeadlinesTool struct {
	tracker *DeadlineTracker
}

func NewViewDeadlinesTool(tracker *DeadlineTracker) *ViewDeadlinesTool {
	return &ViewDeadlinesTool{tracker: tracker}
}

func (t *ViewDeadlinesTool) Name() string { return "view_deadlines" }
func (t *ViewDeadlinesTool) Description() string {
	return "View the student's upcoming pending deadlines, assignments, and exams."
}

func (t *ViewDeadlinesTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (t *ViewDeadlinesTool) Execute(ctx context.Context, params map[string]any) *tools.ToolResult {
	upcoming, err := t.tracker.GetUpcoming()
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to fetch deadlines: %v", err))
	}

	if len(upcoming) == 0 {
		return tools.SuccessResult("No upcoming deadlines found in the tracker. You are free for now!", "No deadlines.")
	}

	msg := "📅 **Upcoming Deadlines:**\n\n"
	for _, d := range upcoming {
		daysOut := int(time.Until(d.DueDate).Hours() / 24)
		msg += fmt.Sprintf("- **%s** (Due in %d days)\n", d.Title, daysOut)
	}

	return tools.SuccessResult(msg, "Retrieved deadlines list.")
}
