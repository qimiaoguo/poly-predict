package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/poly-predict/backend/pkg/config"
	"github.com/poly-predict/backend/pkg/db"
	"github.com/poly-predict/backend/services/scraper/internal/polymarket"
	"github.com/poly-predict/backend/services/scraper/internal/scheduler"
	"github.com/poly-predict/backend/services/scraper/internal/syncer"
)

func main() {
	// Configure zerolog.
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Str("service", "scraper").
		Logger()

	log.Info().Msg("starting scraper service")

	// Load configuration.
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	log.Info().
		Str("environment", cfg.Environment).
		Msg("config loaded")

	// Initialize database connection pool.
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize database pool")
	}
	defer db.Close(pool)

	log.Info().Msg("database connection established")

	// Create API clients and syncer.
	gammaClient := polymarket.NewGammaClient()
	clobClient := polymarket.NewCLOBClient()
	syncService := syncer.New(pool, gammaClient, clobClient)

	// Run an initial sync immediately.
	log.Info().Msg("running initial sync")
	if err := syncService.SyncAll(ctx); err != nil {
		log.Error().Err(err).Msg("initial sync failed")
	}

	// Set up cron scheduler.
	sched := scheduler.New(syncService)
	sched.Start()

	// Block until shutdown signal.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	log.Info().Str("signal", sig.String()).Msg("received shutdown signal")

	// Graceful shutdown.
	sched.Stop()
	db.Close(pool)

	log.Info().Msg("scraper service stopped")
}
