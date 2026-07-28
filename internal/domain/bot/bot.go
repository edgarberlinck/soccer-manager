package bot

import (
	"manager/game/internal/domain/player"
	"manager/game/internal/domain/training"
	"math/rand"
)

type Strategy string

const (
	Conservador Strategy = "conservador"
	Equilibrado Strategy = "equilibrado"
	Agressivo   Strategy = "agressivo"
)

type Bot struct {
	ID       string
	Username string
	Strategy Strategy
	ClubID   string
}

// DecideTraining escolhe um treino baseado na estratégia do bot
func (b *Bot) DecideTraining(p player.Player) training.TrainingSession {
	trainingType := b.selectTrainingType()
	intensity := b.selectIntensity(p)
	
	return training.TrainingSession{
		Type:     trainingType,
		Intensity: intensity,
		Duration: trainingType.BaseDuration(),
	}
}

func (b *Bot) selectTrainingType() training.TrainingType {
	types := []training.TrainingType{
		training.Finishing,
		training.Passing,
		training.Dribbling,
		training.Shooting,
		training.Speed,
		training.Strength,
		training.Stamina,
		training.Goalkeeping,
	}
	
	return types[rand.Intn(len(types))]
}

func (b *Bot) selectIntensity(p player.Player) training.Intensity {
	switch b.Strategy {
	case Conservador:
		// Prioriza Soft e Medium
		if p.Attributes.FisicalStatus < 50 {
			return training.Soft
		}
		if rand.Float32() < 0.7 {
			return training.Soft
		}
		return training.Medium
		
	case Equilibrado:
		// Mix balanceado
		if p.Attributes.FisicalStatus < 40 {
			return training.Soft
		}
		r := rand.Float32()
		if r < 0.33 {
			return training.Soft
		} else if r < 0.66 {
			return training.Medium
		}
		return training.Intense
		
	case Agressivo:
		// Prioriza Intense e Medium
		if p.Attributes.FisicalStatus < 30 {
			return training.Medium
		}
		if rand.Float32() < 0.6 {
			return training.Intense
		}
		return training.Medium
		
	default:
		return training.Medium
	}
}

// SelectBestPlayers seleciona os melhores jogadores para uma partida
func (b *Bot) SelectBestPlayers(players []player.Player, count int) []player.Player {
	if len(players) <= count {
		return players
	}
	
	// Ordena jogadores por overall (média de atributos)
	sorted := make([]player.Player, len(players))
	copy(sorted, players)
	
	// Bubble sort simples por overall
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if calculateOverall(sorted[j]) < calculateOverall(sorted[j+1]) {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	
	return sorted[:count]
}

func calculateOverall(p player.Player) float64 {
	attrs := p.Attributes
	sum := attrs.Pace + attrs.Passing + attrs.Shooting + attrs.Fisico +
		attrs.FisicalStatus + attrs.Cabeceio + attrs.Cruzamento +
		attrs.Habilidade + attrs.Finalizacao + attrs.Dominio
	return float64(sum) / 10.0
}

// ShouldRest determina se o jogador deve descansar baseado na estratégia
func (b *Bot) ShouldRest(p player.Player) bool {
	switch b.Strategy {
	case Conservador:
		return p.Attributes.FisicalStatus < 60
	case Equilibrado:
		return p.Attributes.FisicalStatus < 40
	case Agressivo:
		return p.Attributes.FisicalStatus < 25
	default:
		return p.Attributes.FisicalStatus < 40
	}
}

// SelectStartingEleven seleciona os 11 titulares para uma partida
// Prioriza jogadores com melhor overall e condição física adequada
func (b *Bot) SelectStartingEleven(allPlayers []player.Player) []player.Player {
	// Filtrar jogadores muito cansados (exceto se não houver escolha)
	availablePlayers := make([]player.Player, 0)
	for _, p := range allPlayers {
		// Conservadores são mais rigorosos
		minStatus := 30
		if b.Strategy == Conservador {
			minStatus = 50
		} else if b.Strategy == Equilibrado {
			minStatus = 40
		}
		
		if p.Attributes.FisicalStatus >= minStatus {
			availablePlayers = append(availablePlayers, p)
		}
	}
	
	// Se não tiver jogadores suficientes, usa todos
	if len(availablePlayers) < 11 {
		availablePlayers = allPlayers
	}
	
	// Seleciona os 11 melhores
	return b.SelectBestPlayers(availablePlayers, 11)
}

// WillAcceptMatch determina se o bot aceita uma partida
// Bots sempre aceitam, mas podem ter lógica futura (ex: tempo desde última partida)
func (b *Bot) WillAcceptMatch() bool {
	// Por enquanto, bots sempre aceitam
	return true
}
