package services

import (
	"context"
	"errors"
	"log"
	"math/rand"
	repository "manager/game/internal/infrastructure/database/generated"

	"github.com/google/uuid"
)

type MatchmakingService struct {
	queries *repository.Queries
}

func NewMatchmakingService(queries *repository.Queries) *MatchmakingService {
	return &MatchmakingService{
		queries: queries,
	}
}

// FindOpponent encontra um oponente para um jogador
// Se não houver jogadores humanos disponíveis, retorna um bot
func (s *MatchmakingService) FindOpponent(ctx context.Context, playerClubID uuid.UUID) (uuid.UUID, bool, error) {
	// Por enquanto, sempre retorna um bot aleatório
	// Futuramente pode implementar matchmaking por ranking, etc
	botClub, err := s.queries.GetRandomBotClub(ctx)
	if err != nil {
		return uuid.Nil, false, err
	}
	
	// Retorna o clube do bot (true = é bot)
	return botClub.ID, true, nil
}

// CreateMatchWithBot cria uma partida entre um jogador e um bot
func (s *MatchmakingService) CreateMatchWithBot(ctx context.Context, playerClubID uuid.UUID) (uuid.UUID, error) {
	botClubID, isBot, err := s.FindOpponent(ctx, playerClubID)
	if err != nil {
		return uuid.Nil, err
	}
	
	if !isBot {
		return uuid.Nil, errors.New("oponente não é um bot")
	}
	
	matchID := uuid.New()
	randomSeed := rand.Int63()
	
	match, err := s.queries.CreateMatch(ctx, repository.CreateMatchParams{
		ID:           matchID,
		HomeClubID:   playerClubID,
		AwayClubID:   botClubID,
		RandomSeed:   randomSeed,
	})
	if err != nil {
		return uuid.Nil, err
	}
	
	log.Printf("Partida criada: %s vs %s (bot)", playerClubID, botClubID)
	
	return match.ID, nil
}

// AutoAcceptBotMatches processa partidas pendentes onde um bot está envolvido
// e automaticamente as aceita/escala o time
func (s *MatchmakingService) AutoAcceptBotMatches(ctx context.Context) error {
	// Busca partidas pendentes
	pendingMatches, err := s.queries.GetPendingMatches(ctx, 50)
	if err != nil {
		return err
	}
	
	log.Printf("Processando %d partidas pendentes para bots...", len(pendingMatches))
	
	for _, match := range pendingMatches {
		// Verifica se algum dos times é bot
		homeIsBot, err := s.isClubBot(ctx, match.HomeClubID)
		if err != nil {
			log.Printf("Erro ao verificar se home é bot: %v", err)
			continue
		}
		
		awayIsBot, err := s.isClubBot(ctx, match.AwayClubID)
		if err != nil {
			log.Printf("Erro ao verificar se away é bot: %v", err)
			continue
		}
		
		// Se pelo menos um é bot, automaticamente aceita/escala
		if homeIsBot || awayIsBot {
			// Aqui você implementaria a lógica de escalar o time do bot
			// Por enquanto apenas logamos
			log.Printf("Match %s: Home=%v, Away=%v (bot flags: %v, %v)", 
				match.ID, match.HomeClubID, match.AwayClubID, homeIsBot, awayIsBot)
			
			// TODO: Implementar escalação automática do time bot
			// TODO: Marcar partida como pronta para começar
		}
	}
	
	return nil
}

func (s *MatchmakingService) isClubBot(ctx context.Context, clubID uuid.UUID) (bool, error) {
	// Usa a query otimizada
	isBot, err := s.queries.IsClubOwnedByBot(ctx, clubID)
	if err != nil {
		return false, err
	}
	
	return isBot, nil
}
