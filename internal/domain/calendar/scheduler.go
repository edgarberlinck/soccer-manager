package calendar

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
)

// MatchDistribution representa uma partida entre dois clubes
type MatchDistribution struct {
	HomeClubID uuid.UUID
	AwayClubID uuid.UUID
	Round      int
}

// RoundRobinScheduler cria um calendário de partidas usando algoritmo round-robin
type RoundRobinScheduler struct {
	Clubs            []uuid.UUID
	TwoLegs          bool // Ida e volta
	ShuffleFixtures  bool // Embaralha ordem das partidas
}

// GenerateMatches gera todas as partidas usando algoritmo round-robin
// Garante que todos os times jogam contra todos
func (s *RoundRobinScheduler) GenerateMatches() ([]MatchDistribution, error) {
	if len(s.Clubs) < 2 {
		return nil, errors.New("need at least 2 clubs to generate matches")
	}

	numClubs := len(s.Clubs)
	
	// Se número ímpar de clubes, adiciona um "bye" (descanso)
	clubs := make([]uuid.UUID, numClubs)
	copy(clubs, s.Clubs)
	hasBye := false
	
	if numClubs%2 != 0 {
		clubs = append(clubs, uuid.Nil) // Nil representa "bye"
		hasBye = true
		numClubs++
	}

	matches := make([]MatchDistribution, 0)
	numRounds := numClubs - 1

	// Algoritmo Round-Robin clássico
	for round := 0; round < numRounds; round++ {
		roundMatches := s.generateRound(clubs, round, hasBye)
		matches = append(matches, roundMatches...)
	}

	// Se two-legs, gera o returno (ida e volta)
	if s.TwoLegs {
		returnMatches := make([]MatchDistribution, 0, len(matches))
		maxRound := numRounds
		for _, match := range matches {
			returnMatches = append(returnMatches, MatchDistribution{
				HomeClubID: match.AwayClubID,
				AwayClubID: match.HomeClubID,
				Round:      match.Round + maxRound,
			})
		}
		matches = append(matches, returnMatches...)
	}

	// Embaralha se necessário
	if s.ShuffleFixtures {
		rand.Shuffle(len(matches), func(i, j int) {
			matches[i], matches[j] = matches[j], matches[i]
		})
		// Renumera rounds após embaralhar
		for i := range matches {
			matches[i].Round = i / (numClubs / 2) + 1
		}
	}

	return matches, nil
}

func (s *RoundRobinScheduler) generateRound(clubs []uuid.UUID, round int, hasBye bool) []MatchDistribution {
	numClubs := len(clubs)
	matches := make([]MatchDistribution, 0, numClubs/2)

	// Array rotativo para o algoritmo
	positions := make([]int, numClubs)
	for i := range positions {
		positions[i] = i
	}

	// Rotaciona posições (exceto a primeira)
	rotateCount := round
	for r := 0; r < rotateCount; r++ {
		temp := positions[numClubs-1]
		for i := numClubs - 1; i > 1; i-- {
			positions[i] = positions[i-1]
		}
		positions[1] = temp
	}

	// Gera as partidas da rodada
	for i := 0; i < numClubs/2; i++ {
		home := clubs[positions[i]]
		away := clubs[positions[numClubs-1-i]]

		// Pula se algum time é "bye"
		if hasBye && (home == uuid.Nil || away == uuid.Nil) {
			continue
		}

		matches = append(matches, MatchDistribution{
			HomeClubID: home,
			AwayClubID: away,
			Round:      round + 1,
		})
	}

	return matches
}

// MatchScheduleBuilder constrói um calendário completo com ticks
type MatchScheduleBuilder struct {
	Matches           []MatchDistribution
	SeasonStartTick   int
	SeasonEndTick     int
	MatchDurationTicks int
	BreakBetweenRounds int // Ticks de intervalo entre rodadas
}

// Build distribui as partidas no tempo disponível
func (b *MatchScheduleBuilder) Build() ([]MatchSlot, error) {
	if len(b.Matches) == 0 {
		return []MatchSlot{}, nil
	}

	if b.MatchDurationTicks <= 0 {
		return nil, errors.New("match_duration_ticks must be > 0")
	}

	if b.SeasonStartTick >= b.SeasonEndTick {
		return nil, errors.New("season_end_tick must be > season_start_tick")
	}

	// Agrupa partidas por rodada
	rounds := make(map[int][]MatchDistribution)
	maxRound := 0
	for _, match := range b.Matches {
		rounds[match.Round] = append(rounds[match.Round], match)
		if match.Round > maxRound {
			maxRound = match.Round
		}
	}

	// Calcula tempo disponível
	totalTicks := b.SeasonEndTick - b.SeasonStartTick
	ticksPerRound := b.MatchDurationTicks + b.BreakBetweenRounds
	requiredTicks := maxRound * ticksPerRound

	if requiredTicks > totalTicks {
		return nil, fmt.Errorf("not enough ticks: need %d, have %d", requiredTicks, totalTicks)
	}

	// Distribui rodadas uniformemente
	slots := make([]MatchSlot, 0, len(b.Matches))
	currentTick := b.SeasonStartTick

	for round := 1; round <= maxRound; round++ {
		roundMatches := rounds[round]
		
		for _, match := range roundMatches {
			slots = append(slots, MatchSlot{
				MatchID:        uuid.New(),
				Kind:           EntryChampionshipMatch,
				Title:          fmt.Sprintf("Round %d", round),
				StartTick:      currentTick,
				EndTick:        currentTick + b.MatchDurationTicks - 1,
				OpponentClubID: match.AwayClubID,
			})
		}

		currentTick += ticksPerRound
	}

	return slots, nil
}

// BuildForClub constrói o calendário para um clube específico
func (b *MatchScheduleBuilder) BuildForClub(clubID uuid.UUID) ([]MatchSlot, error) {
	allSlots, err := b.Build()
	if err != nil {
		return nil, err
	}

	// Filtra apenas partidas deste clube
	clubSlots := make([]MatchSlot, 0)
	matchIndex := 0

	for _, match := range b.Matches {
		if match.HomeClubID == clubID {
			if matchIndex < len(allSlots) {
				slot := allSlots[matchIndex]
				slot.OpponentClubID = match.AwayClubID
				clubSlots = append(clubSlots, slot)
			}
			matchIndex++
		} else if match.AwayClubID == clubID {
			if matchIndex < len(allSlots) {
				slot := allSlots[matchIndex]
				slot.OpponentClubID = match.HomeClubID
				clubSlots = append(clubSlots, slot)
			}
			matchIndex++
		}
	}

	return clubSlots, nil
}

// ValidateDistribution verifica se a distribuição é válida
func ValidateDistribution(matches []MatchDistribution, clubs []uuid.UUID) error {
	if len(matches) == 0 {
		return errors.New("no matches to validate")
	}

	clubSet := make(map[uuid.UUID]bool)
	for _, club := range clubs {
		clubSet[club] = true
	}

	// Verifica se todos os clubes nas partidas existem
	for i, match := range matches {
		if !clubSet[match.HomeClubID] {
			return fmt.Errorf("match %d: home club %s not in clubs list", i, match.HomeClubID)
		}
		if !clubSet[match.AwayClubID] {
			return fmt.Errorf("match %d: away club %s not in clubs list", i, match.AwayClubID)
		}
		if match.HomeClubID == match.AwayClubID {
			return fmt.Errorf("match %d: club cannot play against itself", i)
		}
		if match.Round < 1 {
			return fmt.Errorf("match %d: round must be >= 1", i)
		}
	}

	// Conta partidas por clube
	matchCount := make(map[uuid.UUID]int)
	for _, match := range matches {
		matchCount[match.HomeClubID]++
		matchCount[match.AwayClubID]++
	}

	// Verifica se todos os clubes da lista jogam
	for _, club := range clubs {
		if matchCount[club] == 0 {
			return fmt.Errorf("club %s has 0 matches", club)
		}
	}

	// Verifica se todos jogam o mesmo número de partidas
	expectedMatches := -1
	for clubID, count := range matchCount {
		if expectedMatches == -1 {
			expectedMatches = count
		} else if count != expectedMatches {
			return fmt.Errorf("club %s has %d matches, expected %d", clubID, count, expectedMatches)
		}
	}

	return nil
}
