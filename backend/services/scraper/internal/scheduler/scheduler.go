package scheduler

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/poly-predict/backend/services/scraper/internal/syncer"
)

// Scheduler wraps a cron scheduler that periodically triggers market syncs.
type Scheduler struct {
	cron   *cron.Cron
	syncer *syncer.Syncer
}

// New creates a new Scheduler.
func New(s *syncer.Syncer) *Scheduler {
	return &Scheduler{
		cron:   cron.New(),
		syncer: s,
	}
}

// Start adds the sync job on a 30-minute interval and starts the cron scheduler.
func (s *Scheduler) Start() {
	_, err := s.cron.AddFunc("@every 30m", func() {
		log.Info().Msg("cron triggered: starting scheduled sync")
		ctx := context.Background()
		if err := s.syncer.SyncAll(ctx); err != nil {
			log.Error().Err(err).Msg("scheduled sync failed")
		}
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to add cron job")
	}

	s.cron.Start()
	log.Info().Msg("cron scheduler started with @every 30m schedule")
}

// Stop gracefully stops the cron scheduler, waiting for running jobs to finish.
func (s *Scheduler) Stop() {
	log.Info().Msg("stopping cron scheduler")
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Info().Msg("cron scheduler stopped")
}
