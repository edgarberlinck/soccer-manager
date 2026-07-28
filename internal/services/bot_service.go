package services

import (
	"context"
	"log"
	"manager/game/internal/domain/bot"
	"manager/game/internal/domain/player"
	"manager/game/internal/domain/training"
	repository "manager/game/internal/infrastructure/database/generated"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type BotService struct {
	queries *repository.Queries
}

func NewBotService(queries *repository.Queries) *BotService {
	return &BotService{
		queries: queries,
	}
}

// RunBotActions executa ações automáticas para todos os bots
func (s *BotService) RunBotActions(ctx context.Context) error {
	bots, err := s.queries.GetAllBots(ctx)
	if err != nil {
		return err
	}

	log.Printf("Executando ações para %d bots...", len(bots))

	for _, botUser := range bots {
		if err := s.executeBotActions(ctx, botUser); err != nil {
			log.Printf("Erro ao executar ações do bot %s: %v", botUser.Username, err)
		}
	}

	return nil
}

func (s *BotService) executeBotActions(ctx context.Context, botUser repository.User) error {
	// Buscar clube do bot
	clubs, err := s.queries.GetUserClubs(ctx, botUser.ID)
	if err != nil || len(clubs) == 0 {
		return err
	}

	club := clubs[0]
	
	// Criar instância do bot
	botInstance := bot.Bot{
		ID:       botUser.ID.String(),
		Username: botUser.Username,
		Strategy: bot.Strategy(botUser.BotStrategy.String),
		ClubID:   club.ID.String(),
	}

	// Buscar jogadores do clube
	players, err := s.queries.GetPlayersByClubId(ctx, uuid.NullUUID{
		UUID:  club.ID,
		Valid: true,
	})
	if err != nil {
		return err
	}

	// Treinar jogadores aleatoriamente (não todos de uma vez)
	numToTrain := rand.Intn(len(players)/3 + 1)
	if numToTrain > 5 {
		numToTrain = 5 // Máximo 5 jogadores por rodada
	}

	shuffled := make([]repository.Player, len(players))
	copy(shuffled, players)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	for i := 0; i < numToTrain && i < len(shuffled); i++ {
		dbPlayer := shuffled[i]
		
		// Converter para domain player
		domainPlayer := player.Player{
			Id:       dbPlayer.ID.String(),
			Name:     dbPlayer.Name,
			Age:      int(dbPlayer.Age),
			Position: dbPlayer.Position,
			Attributes: player.Attributes{
				Pace:          int(dbPlayer.Pace),
				Passing:       int(dbPlayer.Passing),
				Shooting:      int(dbPlayer.Shooting),
				Fisico:        int(dbPlayer.Fisico),
				FisicalStatus: int(dbPlayer.FisicalStatus),
				Cabeceio:      int(dbPlayer.Cabeceio),
				Cruzamento:    int(dbPlayer.Cruzamento),
				Habilidade:    int(dbPlayer.Habilidade),
				Finalizacao:   int(dbPlayer.Finalizacao),
				Dominio:       int(dbPlayer.Dominio),
			},
		}

		// Verificar se deve descansar
		if botInstance.ShouldRest(domainPlayer) {
			// Descanso: recupera status físico
			newStatus := training.ClampPhysicalStatus(domainPlayer.Attributes.FisicalStatus + rand.Intn(15) + 10)
			_, err = s.queries.UpdatePlayerPhysicalStatus(ctx, repository.UpdatePlayerPhysicalStatusParams{
				ID:            dbPlayer.ID,
				FisicalStatus: int16(newStatus),
			})
			if err != nil {
				log.Printf("Erro ao atualizar status físico: %v", err)
			}
			continue
		}

		// Decidir treino
		session := botInstance.DecideTraining(domainPlayer)
		
		// Simular treino
		randomFactor := 0.8 + rand.Float64()*0.4 // 0.8 a 1.2
		gain := training.ComputeTrainingGain(session, domainPlayer, randomFactor)
		fatigue := training.ComputeFatigueCost(session)
		
		// Aplicar ganhos
		updatedAttrs := session.Type.ApplyGain(domainPlayer.Attributes, gain)
		updatedAttrs.FisicalStatus = training.ClampPhysicalStatus(updatedAttrs.FisicalStatus - fatigue)
		
		// Atualizar no banco
		_, err = s.queries.UpdatePlayerAttributes(ctx, repository.UpdatePlayerAttributesParams{
			ID:            dbPlayer.ID,
			Pace:          int16(updatedAttrs.Pace),
			Passing:       int16(updatedAttrs.Passing),
			Shooting:      int16(updatedAttrs.Shooting),
			Altura:        dbPlayer.Altura,
			Peso:          dbPlayer.Peso,
			Impulso:       dbPlayer.Impulso,
			Explosao:      dbPlayer.Explosao,
			Fisico:        int16(updatedAttrs.Fisico),
			FisicalStatus: int16(updatedAttrs.FisicalStatus),
			Cabeceio:      int16(updatedAttrs.Cabeceio),
			Cruzamento:    int16(updatedAttrs.Cruzamento),
			Habilidade:    int16(updatedAttrs.Habilidade),
			Finalizacao:   int16(updatedAttrs.Finalizacao),
			Dominio:       int16(updatedAttrs.Dominio),
			Temperamento:  dbPlayer.Temperamento,
		})
		if err != nil {
			log.Printf("Erro ao atualizar atributos do jogador: %v", err)
		}
	}

	return nil
}

// StartBotScheduler inicia um scheduler que executa ações dos bots periodicamente
func (s *BotService) StartBotScheduler(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Bot scheduler iniciado (intervalo: %v)", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Bot scheduler encerrado")
			return
		case <-ticker.C:
			if err := s.RunBotActions(ctx); err != nil {
				log.Printf("Erro ao executar ações dos bots: %v", err)
			}
		}
	}
}
