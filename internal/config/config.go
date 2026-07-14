package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL               string
	Port                      string
	AppBaseURL                string
	CORSAllowedOrigins        []string
	AuthJWTSecret             string
	AuthJWTExpirationMinutes  int
	AuthVerifyTokenTTLMinutes int
	ResendAPIKey              string
	ResendFromEmail           string
	SimulationTickSeconds     int
	SimulationTickCron        string
	SimulationMaxParallel     int
	SimulationMatchBatchSize  int
	SimulationWorkerPoolSize  int
	SimulationQueueSize       int
	CalendarTickSeconds       int
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:               os.Getenv("DATABASE_URL"),
		Port:                      getEnvOrDefault("PORT", "8080"),
		AppBaseURL:                getEnvOrDefault("APP_BASE_URL", "http://localhost:8080"),
		CORSAllowedOrigins:        getEnvCSVOrDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000,http://localhost:5173,http://127.0.0.1:5173"),
		AuthJWTSecret:             os.Getenv("AUTH_JWT_SECRET"),
		AuthJWTExpirationMinutes:  getEnvAsIntOrDefault("AUTH_JWT_EXPIRATION_MINUTES", 60),
		AuthVerifyTokenTTLMinutes: getEnvAsIntOrDefault("AUTH_VERIFY_TOKEN_TTL_MINUTES", 1440),
		ResendAPIKey:              os.Getenv("RESEND_API_KEY"),
		ResendFromEmail:           os.Getenv("RESEND_FROM_EMAIL"),
		SimulationTickSeconds:     getEnvAsIntOrDefault("SIMULATION_TICK_SECONDS", 5),
		SimulationTickCron:        os.Getenv("SIMULATION_TICK_CRON"),
		SimulationMaxParallel:     getEnvAsIntOrDefault("SIMULATION_MAX_PARALLEL", 8),
		SimulationMatchBatchSize:  getEnvAsIntOrDefault("SIMULATION_MATCH_BATCH_SIZE", 128),
		SimulationWorkerPoolSize:  getEnvAsIntOrDefault("SIMULATION_WORKER_POOL_SIZE", 0),
		SimulationQueueSize:       getEnvAsIntOrDefault("SIMULATION_QUEUE_SIZE", 0),
		CalendarTickSeconds:       getEnvAsIntOrDefault("CALENDAR_TICK_SECONDS", 60),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.AuthJWTSecret == "" {
		return Config{}, fmt.Errorf("AUTH_JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnvOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getEnvAsIntOrDefault(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvCSVOrDefault(key, fallback string) []string {
	raw := os.Getenv(key)
	if strings.TrimSpace(raw) == "" {
		raw = fallback
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, item := range parts {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		values = append(values, trimmed)
	}

	return values
}
