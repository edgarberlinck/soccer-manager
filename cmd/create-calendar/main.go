package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"manager/game/internal/config"
	"manager/game/internal/domain/calendar"
	repository "manager/game/internal/infrastructure/database/generated"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	twoLegs := flag.Bool("two-legs", true, "Generate home and away matches")
	shuffle := flag.Bool("shuffle", false, "Shuffle fixture order")
	matchDuration := flag.Int("match-duration", 90, "Match duration in ticks")
	breakBetween := flag.Int("break", 100, "Break between rounds in ticks")
	flag.Parse()

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
	ctx := context.Background()

	log.Println("🗓️  Criando calendário da temporada...")

	// Busca todos os clubes
	log.Println("Buscando clubes...")
	clubs, err := getAllClubs(ctx, queries)
	if err != nil {
		log.Fatalf("Erro ao buscar clubes: %v", err)
	}

	log.Printf("✅ Encontrados %d clubes", len(clubs))

	if len(clubs) < 2 {
		log.Fatal("❌ Necessário pelo menos 2 clubes para criar calendário")
	}

	// Gera distribuição de partidas
	log.Println("Gerando distribuição de partidas (Round-Robin)...")
	scheduler := calendar.RoundRobinScheduler{
		Clubs:           clubs,
		TwoLegs:         *twoLegs,
		ShuffleFixtures: *shuffle,
	}

	matches, err := scheduler.GenerateMatches()
	if err != nil {
		log.Fatalf("Erro ao gerar partidas: %v", err)
	}

	log.Printf("✅ Geradas %d partidas", len(matches))

	// Valida distribuição
	log.Println("Validando distribuição...")
	if err := calendar.ValidateDistribution(matches, clubs); err != nil {
		log.Fatalf("❌ Erro na validação: %v", err)
	}
	log.Println("✅ Distribuição válida")

	// Calcula ticks da temporada (1 ano = ~525600 minutos)
	now := time.Now()
	seasonStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	seasonEnd := time.Date(now.Year(), 12, 31, 23, 59, 59, 0, time.UTC)
	
	startTick := int(seasonStart.Unix() / 60) // minutos desde epoch
	endTick := int(seasonEnd.Unix() / 60)

	log.Printf("Temporada: %s a %s", seasonStart.Format("2006-01-02"), seasonEnd.Format("2006-01-02"))
	log.Printf("Ticks: %d a %d", startTick, endTick)

	// Cria agenda temporal
	log.Println("Distribuindo partidas no tempo...")
	builder := calendar.MatchScheduleBuilder{
		Matches:            matches,
		SeasonStartTick:    startTick,
		SeasonEndTick:      endTick,
		MatchDurationTicks: *matchDuration,
		BreakBetweenRounds: *breakBetween,
	}

	slots, err := builder.Build()
	if err != nil {
		log.Fatalf("❌ Erro ao criar agenda: %v", err)
	}

	log.Printf("✅ Agenda criada com %d slots temporais", len(slots))

	// Salva partidas no banco
	log.Println("Salvando partidas no banco de dados...")
	created := 0
	failed := 0

	for i, match := range matches {
		if i >= len(slots) {
			log.Printf("⚠️  Sem slots suficientes para todas as partidas")
			break
		}

		championshipID := uuid.Nil // TODO: Criar campeonato

		matchID := uuid.New()
		_, err := queries.CreateMatch(ctx, repository.CreateMatchParams{
			ID:             matchID,
			HomeClubID:     match.HomeClubID,
			AwayClubID:     match.AwayClubID,
			ChampionshipID: uuid.NullUUID{UUID: championshipID, Valid: false},
			RandomSeed:     int64(time.Now().UnixNano()),
		})

		if err != nil {
			log.Printf("❌ Erro ao criar partida %d: %v", i+1, err)
			failed++
			continue
		}

		created++
		if created%100 == 0 {
			log.Printf("Progresso: %d partidas criadas...", created)
		}
	}

	log.Printf("\n✅ Calendário criado com sucesso!")
	log.Printf("Partidas criadas: %d", created)
	log.Printf("Falhas: %d", failed)
	
	// Estatísticas
	matchesPerClub := len(matches) / len(clubs)
	rounds := calculateRounds(len(clubs), *twoLegs)
	
	log.Printf("\n📊 Estatísticas:")
	log.Printf("- Clubes: %d", len(clubs))
	log.Printf("- Partidas por clube: ~%d", matchesPerClub)
	log.Printf("- Total de rodadas: %d", rounds)
	log.Printf("- Duração da partida: %d ticks", *matchDuration)
	log.Printf("- Intervalo entre rodadas: %d ticks", *breakBetween)
	
	if *twoLegs {
		log.Printf("- Formato: Ida e volta ✅")
	} else {
		log.Printf("- Formato: Turno único")
	}
}

func getAllClubs(ctx context.Context, queries *repository.Queries) ([]uuid.UUID, error) {
	// Busca clubes de bots
	botClubs, err := queries.GetBotClubs(ctx)
	if err != nil {
		return nil, err
	}

	clubs := make([]uuid.UUID, len(botClubs))
	for i, club := range botClubs {
		clubs[i] = club.ID
	}

	return clubs, nil
}

func calculateRounds(numClubs int, twoLegs bool) int {
	rounds := numClubs - 1
	if numClubs%2 != 0 {
		rounds = numClubs
	}
	if twoLegs {
		rounds *= 2
	}
	return rounds
}
