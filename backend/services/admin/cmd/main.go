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
	"github.com/poly-predict/backend/services/admin/internal/auth"
	"github.com/poly-predict/backend/services/admin/internal/handler"
	"github.com/poly-predict/backend/services/admin/internal/repository"
	"github.com/poly-predict/backend/services/admin/internal/service"
)

func main() {
	// Configure zerolog.
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load configuration.
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	// Default port for admin service is 8081.
	if os.Getenv("PORT") == "" {
		cfg.Port = "8081"
	}

	// Initialise database pool.
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialise database pool")
	}
	defer db.Close(pool)

	log.Info().Msg("database connection established")

	// JWT secret.
	jwtSecret := cfg.AdminJWTSecret
	if jwtSecret == "" {
		jwtSecret = "admin-secret-change-me"
	}

	// Auth.
	adminAuth := auth.NewAdminAuth(pool, jwtSecret)

	// Repositories.
	userRepo := repository.NewUserRepository(pool)
	eventRepo := repository.NewEventRepository(pool)
	settlementRepo := repository.NewSettlementRepository(pool)
	dashboardRepo := repository.NewDashboardRepository(pool)

	// Services.
	userSvc := service.NewUserService(userRepo)
	eventSvc := service.NewEventService(eventRepo)
	settlementSvc := service.NewSettlementService(settlementRepo)
	dashboardSvc := service.NewDashboardService(dashboardRepo)

	// Handlers.
	authHandler := handler.NewAuthHandler(adminAuth)
	userHandler := handler.NewUserHandler(userSvc)
	eventHandler := handler.NewEventHandler(eventSvc, settlementSvc)
	settlementHandler := handler.NewSettlementHandler(settlementSvc)
	dashboardHandler := handler.NewDashboardHandler(dashboardSvc)

	// Router.
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Public routes.
	api := router.Group("/api/v1")
	api.POST("/auth/login", authHandler.Login)

	// Protected routes.
	protected := api.Group("")
	protected.Use(adminAuth.Middleware())
	{
		protected.GET("/dashboard", dashboardHandler.GetDashboard)

		protected.GET("/users", userHandler.ListUsers)
		protected.GET("/users/:id", userHandler.GetUser)
		protected.PATCH("/users/:id", userHandler.PatchUser)

		protected.GET("/events", eventHandler.ListEvents)
		protected.PATCH("/events/:id", eventHandler.PatchEvent)
		protected.POST("/events/:id/settle", eventHandler.SettleEvent)

		protected.GET("/settlements", settlementHandler.ListSettlements)
	}

	// Start server with graceful shutdown.
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Info().Str("port", cfg.Port).Msg("admin service starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	// Wait for interrupt signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutting down admin service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("admin service stopped")
}
