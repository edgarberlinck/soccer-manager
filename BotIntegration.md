# Integração de Bots com o Scheduler Principal

## Opção 1: Scheduler Separado (Recomendado)

Execute o bot-scheduler como um processo separado:

```bash
# Terminal 1: API principal
cd api
go run cmd/api/main.go

# Terminal 2: Bot Scheduler
go run cmd/bot-scheduler/main.go
```

**Vantagens:**
- Isolamento de processos
- Fácil de debugar
- Pode ser escalado independentemente
- Não afeta performance da API

## Opção 2: Integrar no Scheduler Principal

Se preferir ter tudo em um único processo, modifique `cmd/api/main.go`:

```go
package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"manager/game/engine"
	"manager/game/internal/api"
	"manager/game/internal/config"
	repository "manager/game/internal/infrastructure/database/generated"
	"manager/game/internal/infrastructure/scheduler"
	"manager/game/internal/services"  // ADICIONAR
	"net/http"
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
	router := api.NewRouter(queries, cfg)
	simulationEngine := engine.New()
	gameScheduler := scheduler.New(simulationEngine, queries, cfg)
	
	// ADICIONAR: Bot Service
	botService := services.NewBotService(queries)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go gameScheduler.Run(ctx)
	
	// ADICIONAR: Bot Scheduler
	go botService.StartBotScheduler(ctx, 30*time.Minute)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// ... resto do código
}
```

## Opção 3: Integrar como Job no Scheduler

Se o scheduler usa cron, adicione um job para bots:

```go
// Em scheduler/scheduler.go ou scheduler/handlers.go

func (s *Scheduler) setupBotJob(ctx context.Context) {
	botService := services.NewBotService(s.queries)
	
	// Executar a cada 30 minutos
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := botService.RunBotActions(ctx); err != nil {
					log.Printf("Erro ao executar ações dos bots: %v", err)
				}
			}
		}
	}()
}

// Chamar no método Run() ou init()
func (s *Scheduler) Run(ctx context.Context) {
	s.setupBotJob(ctx) // ADICIONAR
	
	// ... resto do código
}
```

## Configuração de Intervalo

Você pode adicionar uma configuração no `config.go`:

```go
type Config struct {
	// ... campos existentes
	BotSchedulerIntervalMinutes int `env:"BOT_SCHEDULER_INTERVAL_MINUTES" envDefault:"30"`
}
```

E usar no scheduler:

```go
interval := time.Duration(cfg.BotSchedulerIntervalMinutes) * time.Minute
go botService.StartBotScheduler(ctx, interval)
```

## Desabilitando Bots em Desenvolvimento

Adicione uma flag no config:

```go
type Config struct {
	// ... campos existentes
	EnableBots bool `env:"ENABLE_BOTS" envDefault:"true"`
}
```

```go
if cfg.EnableBots {
	go botService.StartBotScheduler(ctx, 30*time.Minute)
	log.Println("✅ Bot scheduler habilitado")
} else {
	log.Println("⚠️  Bot scheduler desabilitado (ENABLE_BOTS=false)")
}
```

## Monitoramento

Adicione métricas para monitorar bots:

```go
// Em bot_service.go
type BotMetrics struct {
	TotalBots       int
	SuccessfulRuns  int
	FailedRuns      int
	LastRunDuration time.Duration
}

func (s *BotService) GetMetrics() BotMetrics {
	// retornar métricas
}
```

## Logs Estruturados

Use zerolog (que já está no projeto) para logs estruturados:

```go
import "github.com/rs/zerolog/log"

log.Info().
	Int("total_bots", len(bots)).
	Str("strategy", string(strategy)).
	Msg("Executando ações dos bots")
```

## Testes

Crie testes para o bot service:

```go
// bot_service_test.go
func TestBotService_RunBotActions(t *testing.T) {
	// mock queries
	// criar bots de teste
	// executar ações
	// verificar resultados
}
```

## Recomendação Final

Para o MVP, recomendo a **Opção 1** (scheduler separado):
- Mais simples de implementar
- Fácil de debugar
- Não aumenta complexidade da API
- Pode ser desligado sem afetar o jogo

Depois, quando o sistema estiver estável, você pode migrar para a **Opção 2** ou **3** se preferir.
