package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"manager/game/internal/config"
	"manager/game/internal/domain/bot"
	repository "manager/game/internal/infrastructure/database/generated"
	"math/rand"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var (
	clubNames = []string{
		"FC Thunder", "United Stars", "City Wolves", "Athletic Lions",
		"Real Warriors", "Sporting Eagles", "Dynamo Tigers", "Galaxy United",
		"Olympic Force", "Victory United", "Rapid Fire", "Phoenix Rising",
		"Vanguard FC", "Titans United", "Storm Breakers", "Lightning FC",
		"Blaze United", "Thunder Strikers", "Velocity FC", "Apex United",
		"Nova Stars", "Fusion FC", "Quantum United", "Nexus City",
		"Velocity Strikers", "Horizon FC", "Eclipse United", "Spectrum City",
		"Pulse FC", "Matrix United", "Orbit Stars", "Zenith City",
		"Vertex FC", "Catalyst United", "Momentum Stars", "Pulse City",
		"Quantum Strikers", "Nexus FC", "Helix United", "Vortex City",
		"Prime FC", "Elite United", "Alpha Stars", "Sigma City",
		"Delta Strikers", "Omega FC", "Beta United", "Gamma Stars",
		"Zeta City", "Theta FC", "Lambda United", "Epsilon Stars",
	}

	countries = []string{
		"Brasil", "Argentina", "Espanha", "Inglaterra", "Alemanha",
		"França", "Itália", "Portugal", "Holanda", "México",
	}

	continents = []string{
		"América do Sul", "Europa", "América do Norte",
	}
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
	ctx := context.Background()

	log.Println("Criando 2000 bots...")
	
	strategies := []bot.Strategy{bot.Conservador, bot.Equilibrado, bot.Agressivo}
	rand.Seed(time.Now().UnixNano())

	created := 0
	failed := 0

	for i := 0; i < 2000; i++ {
		strategy := strategies[i%3]
		
		userID := uuid.New()
		username := fmt.Sprintf("bot_%s_%d", strategy, i)
		
		// Hash vazio pois bots não fazem login
		passwordHash, err := bcrypt.GenerateFromPassword([]byte("bot"), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Erro ao gerar hash: %v", err)
			failed++
			continue
		}

		// Criar usuário bot
		user, err := queries.CreateBotUser(ctx, repository.CreateBotUserParams{
			ID:           userID,
			Username:     username,
			PasswordHash: string(passwordHash),
			BotStrategy:  sql.NullString{String: string(strategy), Valid: true},
		})
		if err != nil {
			log.Printf("Erro ao criar bot %s: %v", username, err)
			failed++
			continue
		}

		// Criar clube para o bot
		clubName := generateClubName(i)
		country := countries[rand.Intn(len(countries))]
		continent := getContinent(country)
		
		clubID := uuid.New()
		_, err = queries.CreateClub(ctx, repository.CreateClubParams{
			ID:           clubID,
			UserID:       user.ID,
			Name:         clubName,
			ShortName:    sql.NullString{String: getShortName(clubName), Valid: true},
			Abbreviation: sql.NullString{String: getAbbreviation(clubName), Valid: true},
			Continent:    sql.NullString{String: continent, Valid: true},
			Country:      sql.NullString{String: country, Valid: true},
		})
		if err != nil {
			log.Printf("Erro ao criar clube para bot %s: %v", username, err)
			failed++
			continue
		}

		// Criar elenco inicial (15 jogadores) para o bot
		if err := createBotSquad(ctx, queries, clubID); err != nil {
			log.Printf("Erro ao criar elenco para bot %s: %v", username, err)
			// Não falha completamente, apenas loga
		}

		created++
		if created%100 == 0 {
			log.Printf("Progresso: %d bots criados...", created)
		}
	}

	log.Printf("\n✅ Processo concluído!")
	log.Printf("Bots criados com sucesso: %d", created)
	log.Printf("Falhas: %d", failed)
	
	// Estatísticas por estratégia
	conservadores := created / 3
	equilibrados := created / 3
	agressivos := created - (conservadores + equilibrados)
	
	log.Printf("\nDistribuição:")
	log.Printf("- Conservadores: ~%d", conservadores)
	log.Printf("- Equilibrados: ~%d", equilibrados)
	log.Printf("- Agressivos: ~%d", agressivos)
}

func generateClubName(index int) string {
	if index < len(clubNames) {
		return clubNames[index]
	}
	
	prefixes := []string{"FC", "United", "City", "Real", "Sporting", "Athletic"}
	suffixes := []string{"Lions", "Tigers", "Eagles", "Wolves", "Stars", "United", "FC", "City"}
	
	prefix := prefixes[rand.Intn(len(prefixes))]
	suffix := suffixes[rand.Intn(len(suffixes))]
	
	return fmt.Sprintf("%s %s %d", prefix, suffix, index)
}

func getShortName(fullName string) string {
	if len(fullName) > 15 {
		return fullName[:15]
	}
	return fullName
}

func getAbbreviation(fullName string) string {
	// Pega as primeiras 3 letras maiúsculas
	abbr := ""
	count := 0
	for _, char := range fullName {
		if char >= 'A' && char <= 'Z' {
			abbr += string(char)
			count++
			if count >= 3 {
				break
			}
		}
	}
	
	if len(abbr) < 3 {
		for len(abbr) < 3 && len(abbr) < len(fullName) {
			abbr += string(fullName[len(abbr)])
		}
	}
	
	return abbr
}

func getContinent(country string) string {
	switch country {
	case "Brasil", "Argentina", "México":
		if country == "México" {
			return "América do Norte"
		}
		return "América do Sul"
	default:
		return "Europa"
	}
}

func createBotSquad(ctx context.Context, queries *repository.Queries, clubID uuid.UUID) error {
	positions := []string{
		"GK", // 1 goleiro
		"DF", "DF", "DF", "DF", // 4 defensores
		"MF", "MF", "MF", "MF", // 4 meio-campistas
		"FW", "FW", "FW", // 3 atacantes
		"GK", "DF", "MF", // 3 reservas variados
	}
	
	firstNames := []string{
		"João", "Pedro", "Lucas", "Gabriel", "Rafael", "Diego", "Carlos", "André",
		"Fernando", "Miguel", "Bruno", "Ricardo", "Roberto", "Paulo", "Marco",
	}
	
	lastNames := []string{
		"Silva", "Santos", "Oliveira", "Souza", "Costa", "Ferreira", "Rodrigues",
		"Alves", "Pereira", "Lima", "Gomes", "Martins", "Ribeiro", "Carvalho",
	}
	
	for i, position := range positions {
		firstName := firstNames[rand.Intn(len(firstNames))]
		lastName := lastNames[rand.Intn(len(lastNames))]
		playerName := fmt.Sprintf("%s %s", firstName, lastName)
		
		// Atributos aleatórios entre 40-75 (times medianos)
		baseAttr := 40 + rand.Intn(36)
		
		_, err := queries.CreateClubPlayer(ctx, repository.CreateClubPlayerParams{
			ID:            uuid.New(),
			ClubID:        uuid.NullUUID{UUID: clubID, Valid: true},
			Name:          playerName,
			Age:           int32(18 + rand.Intn(15)), // 18-32 anos
			Position:      position,
			Overall:       int16(baseAttr + rand.Intn(10)),
			Potential:     int16(baseAttr + 10 + rand.Intn(20)),
			Tier:          "Common",
			Pace:          int16(baseAttr + rand.Intn(20) - 10),
			Passing:       int16(baseAttr + rand.Intn(20) - 10),
			Shooting:      int16(baseAttr + rand.Intn(20) - 10),
			Altura:        int16(160 + rand.Intn(30)),
			Peso:          int16(60 + rand.Intn(30)),
			Impulso:       int16(baseAttr + rand.Intn(20) - 10),
			Explosao:      int16(baseAttr + rand.Intn(20) - 10),
			Fisico:        int16(baseAttr + rand.Intn(20) - 10),
			FisicalStatus: 100, // Começam descansados
			Cabeceio:      int16(baseAttr + rand.Intn(20) - 10),
			Cruzamento:    int16(baseAttr + rand.Intn(20) - 10),
			Habilidade:    int16(baseAttr + rand.Intn(20) - 10),
			Finalizacao:   int16(baseAttr + rand.Intn(20) - 10),
			Dominio:       int16(baseAttr + rand.Intn(20) - 10),
			Temperamento:  int16(50 + rand.Intn(50)),
		})
		if err != nil {
			return fmt.Errorf("erro ao criar jogador %d: %w", i, err)
		}
	}
	
	return nil
}
