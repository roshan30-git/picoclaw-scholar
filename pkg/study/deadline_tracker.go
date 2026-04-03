package study

import (
	"log"
	"time"

	"github.com/roshan30-git/picoclaw-scholar/pkg/database"
)

// Deadline represents a pending task or exam.
type Deadline struct {
	ID       int64
	Title    string
	DueDate  time.Time
	Status   string
	Source   string
}

// DeadlineTracker manages upcoming submissions and exams.
type DeadlineTracker struct {
	db *database.DB
}

// NewDeadlineTracker creates a new instance.
func NewDeadlineTracker(db *database.DB) *DeadlineTracker {
	return &DeadlineTracker{db: db}
}

// AddDeadline creates a new deadline in the database.
func (dt *DeadlineTracker) AddDeadline(title string, dueDate time.Time, source string) error {
	_, err := dt.db.Conn().Exec(
		`INSERT INTO deadlines (title, due_date, status, source) VALUES (?, ?, 'pending', ?)`,
		title, dueDate, source,
	)
	if err != nil {
		log.Printf("[DeadlineTracker] Failed to add %s: %v", title, err)
		return err
	}
	log.Printf("[DeadlineTracker] Added deadline: %s due on %v", title, dueDate)
	return nil
}

// GetUpcoming returns all pending deadlines ordered by closest date.
func (dt *DeadlineTracker) GetUpcoming() ([]Deadline, error) {
	rows, err := dt.db.Conn().Query(
		`SELECT id, title, due_date, status, source FROM deadlines 
		 WHERE status = 'pending' AND due_date >= ? ORDER BY due_date ASC`,
		time.Now().Add(-24*time.Hour), // Include recently missed exactly by a bit for reminders
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Deadline
	for rows.Next() {
		var d Deadline
		if err := rows.Scan(&d.ID, &d.Title, &d.DueDate, &d.Status, &d.Source); err != nil {
			return nil, err
		}
		results = append(results, d)
	}
	return results, nil
}

// MarkCompleted marks a deadline as done.
func (dt *DeadlineTracker) MarkCompleted(id int64) error {
	_, err := dt.db.Conn().Exec(`UPDATE deadlines SET status = 'completed' WHERE id = ?`, id)
	return err
}
