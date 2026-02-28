package study

import (
	"context"
	"log"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
}

func NewScheduler() *Scheduler {
	return &Scheduler{cron: cron.New()}
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
