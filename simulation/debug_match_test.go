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
	if len(snapshots[10].Field.Matrix) != debugFieldHeight || len(snapshots[10].Field.Matrix[0]) != debugFieldWidth {
		t.Fatalf("expected FE matrix sized %dx%d", debugFieldWidth, debugFieldHeight)
	}
	ballCell := snapshots[10].Field.Matrix[snapshots[10].Field.Ball.Y][snapshots[10].Field.Ball.X]
	if len(ballCell.Occupants) == 0 {
		t.Fatalf("expected matrix occupant metadata for ball cell")
	}
	if snapshots[len(snapshots)-1].Outcome.Tick != 90 {
		t.Fatalf("expected last snapshot at tick 90, got %d", snapshots[len(snapshots)-1].Outcome.Tick)
	}
}

func TestSimulateDebugMatchTracksStaminaAndSubstitutions(t *testing.T) {
	home := append(balancedSquad(), TacticalPlayer{
		Name:     "Banco Forte",
		Role:     "forward",
		Position: "Centro Avante",
		Attributes: balancedSquad()[0].Attributes,
	})
	home[9].Attributes.FisicalStatus = 28
	home[11].Attributes.FisicalStatus = 92
	home[11].Attributes.Finalizacao = 88
	home[11].Attributes.Pace = 84

	snapshots := SimulateDebugMatch(DebugMatchConfig{
		MatchID: uuid.MustParse("99999999-1111-1111-1111-111111111111"),
		Seed:    77,
		Home: DebugTeam{Name: "Casa", ClubID: uuid.MustParse("11111111-2222-2222-2222-222222222222"), Players: home},
		Away: DebugTeam{Name: "Fora", ClubID: uuid.MustParse("33333333-4444-4444-4444-444444444444"), Players: balancedSquad()},
	})

	foundSub := false
	for _, snapshot := range snapshots {
		if snapshot.Field.LowestHomeStamina >= 99 {
			t.Fatalf("expected home stamina to decrease during simulation")
		}
		if len(snapshot.Field.Substitutions) > 0 {
			foundSub = true
			break
		}
	}
	if !foundSub {
		t.Fatalf("expected at least one automatic substitution when bench is available")
	}
}