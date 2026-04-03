package study

import (
	"fmt"
	"time"
)

type CalendarEvent struct {
	Name string
	Date time.Time
	Type string // "holiday", "exam"
}

// CalendarEngine provides contextual awareness of the Indian academic calendar.
type CalendarEngine struct {
	events []CalendarEvent
}

func NewCalendarEngine() *CalendarEngine {
	// Hardcoded GTU semester dates & Indian holidays MVP
	now := time.Now()
	year := now.Year()
	
	// Example fixed dates for the MVP. In a real system, these would use a dynamic holiday API or DB.
	events := []CalendarEvent{
		{"Republic Day", time.Date(year, time.January, 26, 0, 0, 0, 0, time.Local), "holiday"},
		{"Holi", time.Date(year, time.March, 25, 0, 0, 0, 0, time.Local), "holiday"},
		{"Summer Semester Exams (GTU)", time.Date(year, time.May, 15, 0, 0, 0, 0, time.Local), "exam"},
		{"Independence Day", time.Date(year, time.August, 15, 0, 0, 0, 0, time.Local), "holiday"},
		{"Mid-Semester Exams (GTU)", time.Date(year, time.October, 15, 0, 0, 0, 0, time.Local), "exam"},
		{"Diwali Break", time.Date(year, time.November, 12, 0, 0, 0, 0, time.Local), "holiday"},
		{"End-Semester Exams (GTU)", time.Date(year, time.December, 10, 0, 0, 0, 0, time.Local), "exam"},
	}

	return &CalendarEngine{events: events}
}

// GetContext returns a string describing upcoming events within the next 45 days.
func (c *CalendarEngine) GetContext() string {
	now := time.Now()
	var upcoming []CalendarEvent

	for _, e := range c.events {
		// If the event has passed in the current year, check next year.
		evtDate := e.Date
		if evtDate.Before(now.Add(-24 * time.Hour)) {
			evtDate = evtDate.AddDate(1, 0, 0)
		}
		
		days := int(evtDate.Sub(now).Hours() / 24)
		if days >= 0 && days <= 45 {
			upcoming = append(upcoming, CalendarEvent{
				Name: e.Name,
				Date: evtDate,
				Type: e.Type,
			})
		}
	}

	if len(upcoming) == 0 {
		return "No major exams or holidays in the next 45 days."
	}

	ctx := "📅 SYSTEM CALENDAR CONTEXT (Upcoming Academic Events):\n"
	for _, e := range upcoming {
		days := int(e.Date.Sub(now).Hours() / 24)
		if days == 0 {
			ctx += fmt.Sprintf("- %s is TODAY!\n", e.Name)
		} else {
			ctx += fmt.Sprintf("- %s: in %d days\n", e.Name, days)
		}
	}
	return ctx
}
