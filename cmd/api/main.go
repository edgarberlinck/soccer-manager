package main

import (
	"manager/game/internal/api"
	"manager/game/internal/config"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
    godotenv.Load()

	cfg := config.Load()
    
    // db, err := sql.Open("pgx", cfg.DatabaseURL)
    // if err != nil {
    //     panic(err)
    // }

    // e := engine.New()
    // s := scheduler.New(e, db) // I need to pass to parameters (see scheduler.go)
    // _ = s
    router := api.NewRouter()

    http.ListenAndServe(":"+cfg.Port, router)
}