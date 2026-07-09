package simulation

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestCanonicalPositionSupportsPortugueseAliases(t *testing.T) {
	tests := map[string]string{
		"Centro Avante": "center_forward",
		"Volante":       "defensive_midfielder",
		"Ala Esquerda":  "left_winger",
		"Lateral Direito": "right_back",
	}

	for input, want := range tests {
		if got := canonicalPosition(input, "midfielder"); got != want {
			t.Fatalf("expected %s for %q, got %s", want, input, got)
		}
	}
}

func TestSimulateDebugMatchRendersFullMatch(t *testing.T) {
	cfg := DebugMatchConfig{
		MatchID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Seed:    42,
		Home: DebugTeam{
			Name:    "Azul",
			ClubID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Players: balancedSquad(),
		},
		Away: DebugTeam{
			Name:    "Vermelho",
			ClubID:  uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			Players: balancedSquad(),
		},
	}

	snapshots := SimulateDebugMatch(cfg)
	if len(snapshots) != regulationTicks {
		t.Fatalf("expected %d snapshots, got %d", regulationTicks, len(snapshots))
	}

	first := RenderDebugSnapshot(snapshots[0])
	if !strings.Contains(first, "Tick 01") {
		t.Fatalf("expected tick header, got %s", first)
	}
	if !strings.Contains(first, "Legenda:") {
		t.Fatalf("expected legend in render, got %s", first)
	}
	if !strings.Contains(first, "J") || !strings.Contains(first, "B") {
		t.Fatalf("expected referee and linesmen markers in render, got %s", first)
	}
	if len(snapshots[10].Field.Home) != 11 || len(snapshots[10].Field.Away) != 11 {
		t.Fatalf("expected 11 players per team in field snapshot")
	}
	if snapshots[10].Field.BallZone == "" || snapshots[10].Field.Possession == "" {
		t.Fatalf("expected tactical state in snapshot: %+v", snapshots[10].Field)
	}
	if snapshots[len(snapshots)-1].Outcome.Tick != 90 {
		t.Fatalf("expected last snapshot at tick 90, got %d", snapshots[len(snapshots)-1].Outcome.Tick)
	}
}