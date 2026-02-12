package scheduler

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/poly-predict/backend/services/settler/internal/settler"
)

// Scheduler wraps a cron scheduler that periodically triggers the settler.
type Scheduler struct {
	cron    *cron.Cron
	settler *settler.Settler
}

// New creates a Scheduler that will invoke the given Settler on each tick.
func New(s *settler.Settler) *Scheduler {
	return &Scheduler{
		cron:    cron.New(),
		settler: s,
	}
}

// Start registers the settlement job and starts the cron scheduler.
func (s *Scheduler) Start() {
	_, err := s.cron.AddFunc("@every 5m", func() {
		log.Info().Msg("cron triggered settlement cycle")
		if err := s.settler.Run(context.Background()); err != nil {
			log.Error().Err(err).Msg("scheduled settlement cycle failed")
		}
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register cron job")
	}

	s.cron.Start()
}

// Stop gracefully stops the cron scheduler, waiting for any running job to finish.
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Info().Msg("cron scheduler stopped")
}
