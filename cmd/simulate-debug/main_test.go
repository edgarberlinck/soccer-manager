package main
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"manager/game/internal/domain/player"
	"manager/game/simulation"

	"github.com/google/uuid"
)

func TestCreateSimulationReportIncludesCalendarAndScore(t *testing.T) {
	now := time.Date(2026, time.July, 13, 15, 4, 0, 0, time.Local)
	home := simulation.DebugTeam{Name: "Casa", ClubID: uuid.New(), Players: testSquad()}
	away := simulation.DebugTeam{Name: "Fora", ClubID: uuid.New(), Players: testSquad()}
	matchID := uuid.New()
	snapshots := []simulation.DebugSnapshot{{Outcome: simulation.TickOutcome{Tick: 1, HomeScore: 0, AwayScore: 0}}, {Outcome: simulation.TickOutcome{Tick: 90, HomeScore: 2, AwayScore: 1}}}

	report, err := createSimulationReport(now, 99, matchID, home, away, snapshots, 60)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.MatchID != matchID {
		t.Fatalf("unexpected match id: %s", report.MatchID)
	}
	if report.HomeScore != 2 || report.AwayScore != 1 {
		t.Fatalf("unexpected final score %d x %d", report.HomeScore, report.AwayScore)
	}
	if report.Calendar.DayEndTick < report.Calendar.DayStartTick {
		t.Fatalf("invalid day tick range: %d > %d", report.Calendar.DayStartTick, report.Calendar.DayEndTick)
	}
	if len(report.Calendar.Administrative) == 0 {
		t.Fatal("expected transfer window in administrative lane")
	}
	if len(report.Calendar.Sporting) == 0 {
		t.Fatal("expected sporting entries")
	}
}

func TestWriteReportCreatesJsonFile(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "nested", "report.json")
	report := simulationReport{
		Seed:        11,
		GeneratedAt: time.Now().Format(time.RFC3339),
		MatchID:     uuid.New(),
		HomeTeam:    "A",
		AwayTeam:    "B",
		HomeScore:   1,
		AwayScore:   0,
		Snapshots:   []simulation.DebugSnapshot{{Outcome: simulation.TickOutcome{Tick: 90}}},
	}

	if err := writeReport(outPath, report); err != nil {
		t.Fatalf("write report failed: %v", err)
	}

	payload, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read report failed: %v", err)
	}

	var decoded simulationReport
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}
	if decoded.Seed != report.Seed {
		t.Fatalf("unexpected seed in report: %d", decoded.Seed)
	}
}

func testSquad() []simulation.TacticalPlayer {
	players := make([]simulation.TacticalPlayer, 0, 11)
	for i := 0; i < 11; i++ {
		players = append(players, simulation.TacticalPlayer{
			Name:     "P",
			Role:     "midfielder",
			Position: "Meia Central",
			Attributes: player.Attributes{
				Pace:          70,
				Passing:       70,
				Shooting:      70,
				Altura:        178,
				Peso:          75,
				Impulso:       70,
				Explosao:      70,
				Fisico:        70,
				FisicalStatus: 85,
				Cabeceio:      70,
				Cruzamento:    70,
				Habilidade:    70,
				Finalizacao:   70,
				Dominio:       70,
				Temperamento:  70,
			},
		})
	}
	return players
}
