package cron

import (
	"log"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		cron: cron.New(),
	}
}

func (s *Scheduler) Start() {
	s.cron.Start()
	log.Println("Cron scheduler started")
	log.Printf("Cron Info: %+v\n", s.cron.Entries())

}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("Cron scheduler stopped")
}

// Register Jobs That want to run
func (s *Scheduler) RegisterJobs() error {

	_, err := s.cron.AddFunc("59 23 * * *", s.ResetDailyReminders)

	if err != nil {
		return err
	}

	return nil
}
