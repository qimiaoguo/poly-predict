package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/poly-predict/backend/pkg/config"
	"github.com/poly-predict/backend/pkg/db"
	"github.com/poly-predict/backend/services/settler/internal/scheduler"
	"github.com/poly-predict/backend/services/settler/internal/settler"
)

func main() {
	// Configure zerolog for human-friendly console output.
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Str("service", "settler").
		Logger()

	log.Info().Msg("starting settler service")

	// Load configuration from environment / .env file.
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	// Initialise the database connection pool.
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialise database pool")
	}
	defer db.Close(pool)

	log.Info().Msg("database pool initialised")

	// Create the settler and run an initial settlement cycle immediately.
	s := settler.New(pool)

	log.Info().Msg("running initial settlement cycle")
	if err := s.Run(ctx); err != nil {
		log.Error().Err(err).Msg("initial settlement cycle failed")
	}

	// Set up the cron scheduler.
	sched := scheduler.New(s)
	sched.Start()
	defer sched.Stop()

	log.Info().Msg("cron scheduler started – settling every 5 minutes")

	// Block until SIGINT or SIGTERM is received.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	log.Info().Str("signal", sig.String()).Msg("received shutdown signal")
	log.Info().Msg("settler service stopped")
}
