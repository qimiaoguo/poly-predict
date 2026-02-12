package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	DatabaseURL       string
	RedisURL          string
	Port              string
	SupabaseURL       string
	SupabaseJWTSecret string
	AdminJWTSecret    string
	Environment       string
}

// Load reads configuration from a .env file (if present) and environment variables.
func Load() (*Config, error) {
	// Load .env file if it exists; ignore error if the file is missing.
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		RedisURL:          os.Getenv("REDIS_URL"),
		Port:              os.Getenv("PORT"),
		SupabaseURL:       os.Getenv("SUPABASE_URL"),
		SupabaseJWTSecret: os.Getenv("SUPABASE_JWT_SECRET"),
		AdminJWTSecret:    os.Getenv("ADMIN_JWT_SECRET"),
		Environment:       os.Getenv("ENVIRONMENT"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	if cfg.Environment == "" {
		cfg.Environment = "development"
	}

	return cfg, nil
}
