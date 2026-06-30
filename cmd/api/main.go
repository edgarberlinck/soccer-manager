package main

import (
	"database/sql"
	"log"
	"manager/game/internal/api"
	"manager/game/internal/config"
	repository "manager/game/internal/infrastructure/database/generated"
	"net/http"

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

	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
