package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/poly-predict/backend/pkg/config"
	"github.com/poly-predict/backend/pkg/db"
	"github.com/poly-predict/backend/services/api/internal/auth"
	"github.com/poly-predict/backend/services/api/internal/handler"
	"github.com/poly-predict/backend/services/api/internal/repository"
	"github.com/poly-predict/backend/services/api/internal/service"
)

func main() {
	// Set up zerolog.
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load configuration.
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database pool.
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close(pool)

	log.Info().Msg("connected to database")

	// Repositories.
	eventRepo := repository.NewEventRepository(pool)
	betRepo := repository.NewBetRepository(pool)
	userRepo := repository.NewUserRepository(pool)

	// Services.
	eventService := service.NewEventService(eventRepo)
	betService := service.NewBetService(pool, betRepo, userRepo)
	rankingService := service.NewRankingService(pool)

	// Handlers.
	eventHandler := handler.NewEventHandler(eventService)
	betHandler := handler.NewBetHandler(betService)
	userHandler := handler.NewUserHandler(userRepo)
	rankingHandler := handler.NewRankingHandler(rankingService)

	// Auth middleware.
	authMiddleware := auth.NewMiddleware(cfg.SupabaseJWTSecret, cfg.SupabaseURL)

	// Set up Gin router.
	router := gin.Default()

	// CORS configuration.
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes.
	api := router.Group("/api/v1")

	// Public routes.
	events := api.Group("/events")
	events.Use(authMiddleware.OptionalAuth())
	{
		events.GET("", eventHandler.ListEvents)
		events.GET("/:id", eventHandler.GetEvent)
		events.GET("/:id/price-history", eventHandler.GetPriceHistory)
	}

	api.GET("/categories", eventHandler.GetCategories)

	rankings := api.Group("/rankings")
	{
		rankings.GET("", rankingHandler.GetRankings)
	}

	// Authenticated routes.
	authenticated := api.Group("")
	authenticated.Use(authMiddleware.RequireAuth())
	{
		authenticated.POST("/bets", betHandler.PlaceBet)
		authenticated.GET("/bets", betHandler.ListBets)
		authenticated.GET("/bets/:id", betHandler.GetBet)

		authenticated.GET("/users/me", userHandler.GetProfile)
		authenticated.PATCH("/users/me", userHandler.UpdateProfile)
		authenticated.GET("/users/me/transactions", userHandler.GetTransactions)
	}

	// Create HTTP server.
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in a goroutine.
	go func() {
		log.Info().Str("port", cfg.Port).Msg("starting API server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server exited")
}
