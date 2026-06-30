package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL               string
	Port                      string
	AppBaseURL                string
	AuthJWTSecret             string
	AuthJWTExpirationMinutes  int
	AuthVerifyTokenTTLMinutes int
	ResendAPIKey              string
	ResendFromEmail           string
	SimulationTickSeconds     int
	SimulationTickCron        string
	SimulationMaxParallel     int
	SimulationMatchBatchSize  int
}

func Load() Config {
	return Config{
		DatabaseURL:               os.Getenv("DATABASE_URL"),
		Port:                      getEnvOrDefault("PORT", "8080"),
		AppBaseURL:                getEnvOrDefault("APP_BASE_URL", "http://localhost:8080"),
		AuthJWTSecret:             os.Getenv("AUTH_JWT_SECRET"),
		AuthJWTExpirationMinutes:  getEnvAsIntOrDefault("AUTH_JWT_EXPIRATION_MINUTES", 60),
		AuthVerifyTokenTTLMinutes: getEnvAsIntOrDefault("AUTH_VERIFY_TOKEN_TTL_MINUTES", 1440),
		ResendAPIKey:              os.Getenv("RESEND_API_KEY"),
		ResendFromEmail:           os.Getenv("RESEND_FROM_EMAIL"),
		SimulationTickSeconds:     getEnvAsIntOrDefault("SIMULATION_TICK_SECONDS", 5),
		SimulationTickCron:        os.Getenv("SIMULATION_TICK_CRON"),
		SimulationMaxParallel:     getEnvAsIntOrDefault("SIMULATION_MAX_PARALLEL", 8),
		SimulationMatchBatchSize:  getEnvAsIntOrDefault("SIMULATION_MATCH_BATCH_SIZE", 128),
	}
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
