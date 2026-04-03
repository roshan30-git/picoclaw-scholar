package study

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/roshan30-git/picoclaw-scholar/integrations/classroom"
	"github.com/roshan30-git/picoclaw-scholar/pkg/bus"
)

type Scheduler struct {
	cron          *cron.Cron
	tracker       *DeadlineTracker
	cards         *WeeklyCardsGenerator
	bus           *bus.MessageBus
	ownerID       string // JID or ChatID to send reminders
	activeChannel string // "telegram" or "whatsapp"
}

func NewScheduler(tracker *DeadlineTracker, cards *WeeklyCardsGenerator, b *bus.MessageBus, ownerID string, activeChannel string) *Scheduler {
	if activeChannel == "" {
		activeChannel = "whatsapp"
	}
	return &Scheduler{
		cron:          cron.New(),
		tracker:       tracker,
		cards:         cards,
		bus:           b,
		ownerID:       ownerID,
		activeChannel: activeChannel,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.cron.Start()
	log.Println("Scheduler started")
	<-ctx.Done()
	s.cron.Stop()
	log.Println("Scheduler stopped")
}

func (s *Scheduler) ScheduleDailyQuiz(timeSpec string, fn func()) {
	s.cron.AddFunc(timeSpec, fn)
}

// SyncClassroom sets up a cron job to sync deadlines from Google Classroom every 6 hours.
func (s *Scheduler) SyncClassroom(gc *classroom.Client) {
	if gc == nil {
		return
	}
	s.cron.AddFunc("@every 6h", func() {
		log.Println("[Scheduler] Syncing Google Classroom deadlines...")
		courses, err := gc.ListCourses(context.Background())
		if err != nil {
			log.Printf("[Scheduler] Classroom sync failed: %v", err)
			return
		}

		for _, c := range courses {
			assignments, _ := gc.ListAssignments(context.Background(), c.Id)
			for _, a := range assignments {
				if a.DueDate != nil {
					// Convert classroom Date/Time to time.Time
					t := time.Date(int(a.DueDate.Year), time.Month(a.DueDate.Month), int(a.DueDate.Day), 23, 59, 0, 0, time.Local)
					s.tracker.AddDeadline(fmt.Sprintf("[%s] %s", c.Name, a.Title), t, "classroom")
				}
			}
		}
	})
}

// ScheduleReminders sets up a daily morning check for upcoming deadlines.
func (s *Scheduler) ScheduleReminders() {
	s.cron.AddFunc("0 8 * * *", func() { // Every day at 8:00 AM
		upcoming, err := s.tracker.GetUpcoming()
		if err != nil || len(upcoming) == 0 {
			return
		}

		msg := "📅 *__Morning Deadline Briefing__*\n\n"
		count := 0
		for _, d := range upcoming {
			hoursUntil := time.Until(d.DueDate).Hours()
			if hoursUntil > 0 && hoursUntil < 48 { // Due within next 2 days
				msg += fmt.Sprintf("• %s (Due in %.0f hours)\n", d.Title, hoursUntil)
				count++
			}
		}

		if count > 0 {
			s.bus.PublishOutbound(bus.OutboundMessage{
				ChatID:  s.ownerID,
				Content: msg,
				Channel: s.activeChannel,
			})
		}
	})
}

// ScheduleWeeklyCards sets up a cron job to send flash cards every Sunday at 10 AM.
func (s *Scheduler) ScheduleWeeklyCards() {
	if s.cards == nil {
		return
	}
	s.cron.AddFunc("0 10 * * 0", func() {
		s.cards.GenerateAndSend(context.Background())
	})
}
