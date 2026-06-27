package main

import (
	"database/sql"
	"manager/game/engine"
	"manager/game/internal/config"
	"manager/game/internal/infrastructure/scheduler"

	"github.com/joho/godotenv"
)

func main() {
    godotenv.Load()

	cfg := config.Load()
    
    db, err := sql.Open("pgx", cfg.DatabaseURL)
    if err != nil {
        panic(err)
    }

    e := engine.New()
    s := scheduler.New(e, db) // I need to pass to parameters (see scheduler.go)
    _ = s
}