package calendar

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRoundRobinScheduler_GenerateMatches(t *testing.T) {
	t.Run("should generate matches for even number of clubs", func(t *testing.T) {
		clubs := make([]uuid.UUID, 4)
		for i := range clubs {
			clubs[i] = uuid.New()
		}

		scheduler := RoundRobinScheduler{
			Clubs:   clubs,
			TwoLegs: false,
		}

		matches, err := scheduler.GenerateMatches()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// 4 clubes = 6 partidas (cada um joga 3)
		expectedMatches := 6
		if len(matches) != expectedMatches {
			t.Errorf("expected %d matches, got %d", expectedMatches, len(matches))
		}

		// Verifica se todos os clubes jogam
		clubMatches := make(map[uuid.UUID]int)
		for _, match := range matches {
			clubMatches[match.HomeClubID]++
			clubMatches[match.AwayClubID]++
		}

		for _, club := range clubs {
			if clubMatches[club] != 3 {
				t.Errorf("club %s should have 3 matches, got %d", club, clubMatches[club])
			}
		}
	})

	t.Run("should generate matches for odd number of clubs", func(t *testing.T) {
		clubs := make([]uuid.UUID, 5)
		for i := range clubs {
			clubs[i] = uuid.New()
		}

		scheduler := RoundRobinScheduler{
			Clubs:   clubs,
			TwoLegs: false,
		}

		matches, err := scheduler.GenerateMatches()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// 5 clubes = 10 partidas (cada um joga 4)
		expectedMatches := 10
		if len(matches) != expectedMatches {
			t.Errorf("expected %d matches, got %d", expectedMatches, len(matches))
		}

		// Cada clube deve jogar 4 partidas
		clubMatches := make(map[uuid.UUID]int)
		for _, match := range matches {
			clubMatches[match.HomeClubID]++
			clubMatches[match.AwayClubID]++
		}

		for _, club := range clubs {
			if clubMatches[club] != 4 {
				t.Errorf("club %s should have 4 matches, got %d", club, clubMatches[club])
			}
		}
	})

	t.Run("should generate two-legs matches", func(t *testing.T) {
		clubs := make([]uuid.UUID, 4)
		for i := range clubs {
			clubs[i] = uuid.New()
		}

		scheduler := RoundRobinScheduler{
			Clubs:   clubs,
			TwoLegs: true,
		}

		matches, err := scheduler.GenerateMatches()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// 4 clubes com ida e volta = 12 partidas
		expectedMatches := 12
		if len(matches) != expectedMatches {
			t.Errorf("expected %d matches, got %d", expectedMatches, len(matches))
		}

		// Cada clube deve jogar 6 partidas (3 casa, 3 fora)
		clubMatches := make(map[uuid.UUID]int)
		for _, match := range matches {
			clubMatches[match.HomeClubID]++
			clubMatches[match.AwayClubID]++
		}

		for _, club := range clubs {
			if clubMatches[club] != 6 {
				t.Errorf("club %s should have 6 matches, got %d", club, clubMatches[club])
			}
		}
	})

	t.Run("should shuffle fixtures when requested", func(t *testing.T) {
		clubs := make([]uuid.UUID, 4)
		for i := range clubs {
			clubs[i] = uuid.New()
		}

		scheduler := RoundRobinScheduler{
			Clubs:           clubs,
			TwoLegs:         false,
			ShuffleFixtures: true,
		}

		matches, err := scheduler.GenerateMatches()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verifica que rodadas foram renumeradas
		roundSet := make(map[int]bool)
		for _, match := range matches {
			roundSet[match.Round] = true
		}

		// Deve ter 3 rodadas (n-1 rodadas para n times)
		if len(roundSet) != 3 {
			t.Errorf("expected 3 rounds, got %d", len(roundSet))
		}
	})

	t.Run("should error with less than 2 clubs", func(t *testing.T) {
		scheduler := RoundRobinScheduler{
			Clubs: []uuid.UUID{uuid.New()},
		}

		_, err := scheduler.GenerateMatches()
		if err == nil {
			t.Error("expected error with 1 club")
		}
	})

	t.Run("should error with empty clubs", func(t *testing.T) {
		scheduler := RoundRobinScheduler{
			Clubs: []uuid.UUID{},
		}

		_, err := scheduler.GenerateMatches()
		if err == nil {
			t.Error("expected error with empty clubs")
		}
	})

	t.Run("should not have duplicate matches", func(t *testing.T) {
		clubs := make([]uuid.UUID, 6)
		for i := range clubs {
			clubs[i] = uuid.New()
		}

		scheduler := RoundRobinScheduler{
			Clubs:   clubs,
			TwoLegs: false,
		}

		matches, err := scheduler.GenerateMatches()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verifica duplicatas
		matchSet := make(map[string]bool)
		for _, match := range matches {
			key := match.HomeClubID.String() + "-" + match.AwayClubID.String()
			if matchSet[key] {
				t.Errorf("duplicate match found: %s vs %s", match.HomeClubID, match.AwayClubID)
			}
			matchSet[key] = true
		}
	})

	t.Run("should verify clubs never play themselves", func(t *testing.T) {
		clubs := make([]uuid.UUID, 8)
		for i := range clubs {
			clubs[i] = uuid.New()
		}

		scheduler := RoundRobinScheduler{
			Clubs:   clubs,
			TwoLegs: true,
		}

		matches, err := scheduler.GenerateMatches()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, match := range matches {
			if match.HomeClubID == match.AwayClubID {
				t.Errorf("club playing against itself: %s", match.HomeClubID)
			}
		}
	})
}

func TestMatchScheduleBuilder_Build(t *testing.T) {
	t.Run("should build schedule successfully", func(t *testing.T) {
		club1 := uuid.New()
		club2 := uuid.New()

		matches := []MatchDistribution{
			{HomeClubID: club1, AwayClubID: club2, Round: 1},
		}

		builder := MatchScheduleBuilder{
			Matches:            matches,
			SeasonStartTick:    1000,
			SeasonEndTick:      2000,
			MatchDurationTicks: 90,
			BreakBetweenRounds: 10,
		}

		slots, err := builder.Build()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(slots) != 1 {
			t.Errorf("expected 1 slot, got %d", len(slots))
		}

		slot := slots[0]
		if slot.StartTick != 1000 {
			t.Errorf("expected start tick 1000, got %d", slot.StartTick)
		}
		if slot.EndTick != 1089 {
			t.Errorf("expected end tick 1089, got %d", slot.EndTick)
		}
	})

	t.Run("should handle multiple rounds", func(t *testing.T) {
		club1 := uuid.New()
		club2 := uuid.New()
		club3 := uuid.New()
		club4 := uuid.New()

		matches := []MatchDistribution{
			{HomeClubID: club1, AwayClubID: club2, Round: 1},
			{HomeClubID: club3, AwayClubID: club4, Round: 1},
			{HomeClubID: club1, AwayClubID: club3, Round: 2},
			{HomeClubID: club2, AwayClubID: club4, Round: 2},
		}

		builder := MatchScheduleBuilder{
			Matches:            matches,
			SeasonStartTick:    1000,
			SeasonEndTick:      5000,
			MatchDurationTicks: 90,
			BreakBetweenRounds: 100,
		}

		slots, err := builder.Build()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(slots) != 4 {
			t.Errorf("expected 4 slots, got %d", len(slots))
		}

		// Verifica que rodada 2 começa depois da rodada 1 + break
		// Rodada 1: 1000-1089
		// Break: 100
		// Rodada 2: 1190+
		round2Start := 1000 + 90 + 100
		foundRound2 := false
		for _, slot := range slots {
			if slot.Title == "Round 2" {
				if slot.StartTick != round2Start {
					t.Errorf("expected round 2 to start at %d, got %d", round2Start, slot.StartTick)
				}
				foundRound2 = true
			}
		}
		if !foundRound2 {
			t.Error("round 2 not found in slots")
		}
	})

	t.Run("should error with invalid match duration", func(t *testing.T) {
		matches := []MatchDistribution{
			{HomeClubID: uuid.New(), AwayClubID: uuid.New(), Round: 1},
		}

		builder := MatchScheduleBuilder{
			Matches:            matches,
			SeasonStartTick:    1000,
			SeasonEndTick:      2000,
			MatchDurationTicks: 0,
		}

		_, err := builder.Build()
		if err == nil {
			t.Error("expected error with match duration 0")
		}
	})

	t.Run("should error with invalid season ticks", func(t *testing.T) {
		matches := []MatchDistribution{
			{HomeClubID: uuid.New(), AwayClubID: uuid.New(), Round: 1},
		}

		builder := MatchScheduleBuilder{
			Matches:            matches,
			SeasonStartTick:    2000,
			SeasonEndTick:      1000,
			MatchDurationTicks: 90,
		}

		_, err := builder.Build()
		if err == nil {
			t.Error("expected error with end before start")
		}
	})

	t.Run("should error when not enough time", func(t *testing.T) {
		matches := []MatchDistribution{
			{HomeClubID: uuid.New(), AwayClubID: uuid.New(), Round: 1},
			{HomeClubID: uuid.New(), AwayClubID: uuid.New(), Round: 2},
		}

		builder := MatchScheduleBuilder{
			Matches:            matches,
			SeasonStartTick:    1000,
			SeasonEndTick:      1100, // Não tem espaço suficiente
			MatchDurationTicks: 90,
			BreakBetweenRounds: 50,
		}

		_, err := builder.Build()
		if err == nil {
			t.Error("expected error when not enough ticks")
		}
	})

	t.Run("should handle empty matches", func(t *testing.T) {
		builder := MatchScheduleBuilder{
			Matches:            []MatchDistribution{},
			SeasonStartTick:    1000,
			SeasonEndTick:      2000,
			MatchDurationTicks: 90,
		}

		slots, err := builder.Build()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(slots) != 0 {
			t.Errorf("expected 0 slots, got %d", len(slots))
		}
	})
}

func TestMatchScheduleBuilder_BuildForClub(t *testing.T) {
	t.Run("should build schedule for specific club", func(t *testing.T) {
		club1 := uuid.New()
		club2 := uuid.New()
		club3 := uuid.New()

		matches := []MatchDistribution{
			{HomeClubID: club1, AwayClubID: club2, Round: 1},
			{HomeClubID: club3, AwayClubID: club1, Round: 2},
			{HomeClubID: club2, AwayClubID: club3, Round: 3},
		}

		builder := MatchScheduleBuilder{
			Matches:            matches,
			SeasonStartTick:    1000,
			SeasonEndTick:      5000,
			MatchDurationTicks: 90,
			BreakBetweenRounds: 100,
		}

		slots, err := builder.BuildForClub(club1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Club1 joga 2 partidas (rounds 1 e 2)
		if len(slots) != 2 {
			t.Errorf("expected 2 slots for club1, got %d", len(slots))
		}

		// Verifica oponentes
		opponents := make(map[uuid.UUID]bool)
		for _, slot := range slots {
			opponents[slot.OpponentClubID] = true
		}

		if !opponents[club2] || !opponents[club3] {
			t.Error("club1 should play against club2 and club3")
		}
	})

	t.Run("should return empty for club not in matches", func(t *testing.T) {
		club1 := uuid.New()
		club2 := uuid.New()
		club3 := uuid.New()

		matches := []MatchDistribution{
			{HomeClubID: club1, AwayClubID: club2, Round: 1},
		}

		builder := MatchScheduleBuilder{
			Matches:            matches,
			SeasonStartTick:    1000,
			SeasonEndTick:      5000,
			MatchDurationTicks: 90,
			BreakBetweenRounds: 100,
		}

		slots, err := builder.BuildForClub(club3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(slots) != 0 {
			t.Errorf("expected 0 slots for club3, got %d", len(slots))
		}
	})
}

func TestValidateDistribution(t *testing.T) {
	t.Run("should validate correct distribution", func(t *testing.T) {
		clubs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
		
		matches := []MatchDistribution{
			{HomeClubID: clubs[0], AwayClubID: clubs[1], Round: 1},
			{HomeClubID: clubs[1], AwayClubID: clubs[2], Round: 1},
			{HomeClubID: clubs[2], AwayClubID: clubs[0], Round: 2},
		}

		err := ValidateDistribution(matches, clubs)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("should error with empty matches", func(t *testing.T) {
		clubs := []uuid.UUID{uuid.New(), uuid.New()}
		
		err := ValidateDistribution([]MatchDistribution{}, clubs)
		if err == nil {
			t.Error("expected error with empty matches")
		}
	})

	t.Run("should error with club not in list", func(t *testing.T) {
		clubs := []uuid.UUID{uuid.New(), uuid.New()}
		unknownClub := uuid.New()
		
		matches := []MatchDistribution{
			{HomeClubID: clubs[0], AwayClubID: unknownClub, Round: 1},
		}

		err := ValidateDistribution(matches, clubs)
		if err == nil {
			t.Error("expected error with unknown club")
		}
	})

	t.Run("should error when club plays itself", func(t *testing.T) {
		clubs := []uuid.UUID{uuid.New(), uuid.New()}
		
		matches := []MatchDistribution{
			{HomeClubID: clubs[0], AwayClubID: clubs[0], Round: 1},
		}

		err := ValidateDistribution(matches, clubs)
		if err == nil {
			t.Error("expected error when club plays itself")
		}
	})

	t.Run("should error with invalid round", func(t *testing.T) {
		clubs := []uuid.UUID{uuid.New(), uuid.New()}
		
		matches := []MatchDistribution{
			{HomeClubID: clubs[0], AwayClubID: clubs[1], Round: 0},
		}

		err := ValidateDistribution(matches, clubs)
		if err == nil {
			t.Error("expected error with round 0")
		}
	})

	t.Run("should error with unbalanced matches", func(t *testing.T) {
		clubs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
		
		// Club 0 joga 2 vezes, club 1 joga 2 vezes, club 2 joga 0 vezes
		matches := []MatchDistribution{
			{HomeClubID: clubs[0], AwayClubID: clubs[1], Round: 1},
			{HomeClubID: clubs[0], AwayClubID: clubs[1], Round: 2},
		}

		err := ValidateDistribution(matches, clubs)
		if err == nil {
			t.Error("expected error with unbalanced matches")
		}
	})
}

func TestRoundRobinScheduler_generateRound(t *testing.T) {
	t.Run("should generate round without bye", func(t *testing.T) {
		clubs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}
		
		scheduler := RoundRobinScheduler{Clubs: clubs}
		matches := scheduler.generateRound(clubs, 0, false)

		// 4 clubes = 2 partidas por rodada
		if len(matches) != 2 {
			t.Errorf("expected 2 matches, got %d", len(matches))
		}
	})

	t.Run("should skip bye matches", func(t *testing.T) {
		clubs := []uuid.UUID{uuid.New(), uuid.New(), uuid.Nil} // Nil = bye
		
		scheduler := RoundRobinScheduler{Clubs: clubs}
		matches := scheduler.generateRound(clubs, 0, true)

		// Com bye, uma das partidas é pulada
		// Deve ter no máximo 1 partida (2 clubes reais + 1 bye)
		if len(matches) > 1 {
			t.Errorf("expected at most 1 match with bye, got %d", len(matches))
		}

		// Verifica que nenhuma partida tem bye
		for _, match := range matches {
			if match.HomeClubID == uuid.Nil || match.AwayClubID == uuid.Nil {
				t.Error("bye should not appear in matches")
			}
		}
	})

	t.Run("should generate multiple rounds correctly", func(t *testing.T) {
		clubs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}
		
		scheduler := RoundRobinScheduler{Clubs: clubs}
		
		// Gera várias rodadas
		round1 := scheduler.generateRound(clubs, 0, false)
		round2 := scheduler.generateRound(clubs, 1, false)
		round3 := scheduler.generateRound(clubs, 2, false)

		// Cada rodada deve ter 2 partidas
		if len(round1) != 2 || len(round2) != 2 || len(round3) != 2 {
			t.Error("each round should have 2 matches")
		}

		// Partidas devem ser diferentes entre rodadas
		r1Key := round1[0].HomeClubID.String() + round1[0].AwayClubID.String()
		r2Key := round2[0].HomeClubID.String() + round2[0].AwayClubID.String()
		if r1Key == r2Key {
			t.Error("rounds should have different matches")
		}
	})
}

func TestNewServerClock(t *testing.T) {
	t.Run("should use defaults when nil/zero", func(t *testing.T) {
		clock := NewServerClock(nil, 0)
		
		if clock.Location == nil {
			t.Error("location should not be nil")
		}
		if clock.TickDuration == 0 {
			t.Error("tick duration should not be zero")
		}
	})

	t.Run("should use provided values", func(t *testing.T) {
		loc := time.UTC
		duration := 5 * time.Minute
		
		clock := NewServerClock(loc, duration)
		
		if clock.Location != loc {
			t.Error("location mismatch")
		}
		if clock.TickDuration != duration {
			t.Error("duration mismatch")
		}
	})
}

func TestTickAt_EdgeCases(t *testing.T) {
	t.Run("should handle negative tick duration", func(t *testing.T) {
		badClock := ServerClock{
			Location:     time.UTC,
			TickDuration: -1 * time.Minute,
		}
		
		// Deve usar default
		tick := badClock.TickAt(time.Unix(60, 0))
		if tick != 1 {
			t.Errorf("expected tick 1, got %d", tick)
		}
	})
}

func TestBuildSeasonCalendar_EdgeCases(t *testing.T) {
	clubID := uuid.New()
	
	t.Run("should error with nil club ID", func(t *testing.T) {
		input := BuildSeasonCalendarInput{
			ClubID:          uuid.Nil,
			SeasonStartTick: 1000,
			SeasonEndTick:   2000,
		}
		
		_, err := BuildSeasonCalendar(input)
		if err == nil {
			t.Error("expected error with nil club ID")
		}
	})
	
	t.Run("should error with invalid start tick", func(t *testing.T) {
		input := BuildSeasonCalendarInput{
			ClubID:          clubID,
			SeasonStartTick: 0,
			SeasonEndTick:   2000,
		}
		
		_, err := BuildSeasonCalendar(input)
		if err == nil {
			t.Error("expected error with start tick 0")
		}
	})
	
	t.Run("should error with end before start", func(t *testing.T) {
		input := BuildSeasonCalendarInput{
			ClubID:          clubID,
			SeasonStartTick: 2000,
			SeasonEndTick:   1000,
		}
		
		_, err := BuildSeasonCalendar(input)
		if err == nil {
			t.Error("expected error with end before start")
		}
	})
}

func TestPlanMatchSimulation_EdgeCases(t *testing.T) {
	t.Run("should handle negative maxParallel", func(t *testing.T) {
		match := CalendarEntry{Kind: EntryChampionshipMatch}
		plan := PlanMatchSimulation([]CalendarEntry{match}, -1, 10)
		
		if len(plan.Batches) == 0 {
			t.Error("should create at least one batch")
		}
	})
	
	t.Run("should handle zero maxMatchesPerBatch", func(t *testing.T) {
		match := CalendarEntry{Kind: EntryChampionshipMatch}
		plan := PlanMatchSimulation([]CalendarEntry{match}, 5, 0)
		
		if len(plan.Batches) == 0 {
			t.Error("should create at least one batch")
		}
	})
}

func TestBuildForClub_EdgeCases(t *testing.T) {
	t.Run("should propagate build errors", func(t *testing.T) {
		builder := MatchScheduleBuilder{
			Matches:            []MatchDistribution{{HomeClubID: uuid.New(), AwayClubID: uuid.New(), Round: 1}},
			SeasonStartTick:    1000,
			SeasonEndTick:      1000, // Invalid
			MatchDurationTicks: 90,
		}
		
		_, err := builder.BuildForClub(uuid.New())
		if err == nil {
			t.Error("expected error from Build")
		}
	})
}
