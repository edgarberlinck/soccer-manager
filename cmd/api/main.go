package main

import (
	"context"
	"database/sql"
	"log"
	"manager/game/engine"
	"manager/game/internal/api"
	"manager/game/internal/config"
	repository "manager/game/internal/infrastructure/database/generated"
	"manager/game/internal/infrastructure/scheduler"
	"net/http"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if cfg.AuthJWTSecret == "" {
		log.Fatal("AUTH_JWT_SECRET is required")
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	queries := repository.New(db)
	router := api.NewRouter(queries, cfg)
	simulationEngine := engine.New()
	gameScheduler := scheduler.New(simulationEngine, queries, cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go gameScheduler.Run(ctx)

	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
