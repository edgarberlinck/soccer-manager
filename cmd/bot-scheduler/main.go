package main

import (
	"context"
	"database/sql"
	"log"
	"manager/game/internal/config"
	repository "manager/game/internal/infrastructure/database/generated"
	"manager/game/internal/services"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	queries := repository.New(db)
	botService := services.NewBotService(queries)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("🤖 Iniciando Bot Scheduler...")
	log.Println("Os bots vão treinar seus jogadores automaticamente a cada 30 minutos")
	log.Println("Pressione Ctrl+C para encerrar")

	// Executar ações imediatamente
	if err := botService.RunBotActions(ctx); err != nil {
		log.Printf("Erro na execução inicial: %v", err)
	}

	// Iniciar scheduler (a cada 30 minutos)
	botService.StartBotScheduler(ctx, 30*time.Minute)
}
