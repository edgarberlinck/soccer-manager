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

func TestSimulateDebugMatchStoresEventActorAndPlayerStats(t *testing.T) {
	cfg := DebugMatchConfig{
		MatchID: uuid.MustParse("77777777-1111-1111-1111-111111111111"),
		Seed:    21,
		Home: DebugTeam{
			Name:    "Casa",
			ClubID:  uuid.MustParse("11111111-3333-3333-3333-333333333333"),
			Players: namedBalancedSquad("casa"),
		},
		Away: DebugTeam{
			Name:    "Fora",
			ClubID:  uuid.MustParse("22222222-4444-4444-4444-444444444444"),
			Players: namedBalancedSquad("fora"),
		},
	}

	snapshots := SimulateDebugMatch(cfg)
	if len(snapshots) == 0 {
		t.Fatal("expected snapshots")
	}

	foundActor := false
	foundStats := false
	for _, snapshot := range snapshots {
		if snapshot.Field.EventActorID != "" || strings.Contains(snapshot.Outcome.Description, "Jogador:") {
			foundActor = true
		}
		for _, member := range snapshot.Field.HomeSquad {
			if member.Stats.Movement > 0 || member.Stats.Touches > 0 || member.Stats.CorrectTouches > 0 || member.Stats.LongPasses > 0 || member.Stats.ShotsOnGoal > 0 || member.Stats.Fouls > 0 {
				foundStats = true
				break
			}
		}
		if foundActor && foundStats {
			break
		}
	}

	if !foundActor {
		t.Fatal("expected at least one tick with event actor")
	}
	if !foundStats {
		t.Fatal("expected player stats to accumulate during simulation")
	}
}

func namedBalancedSquad(prefix string) []TacticalPlayer {
	squad := balancedSquad()
	for i := range squad {
		squad[i].Name = prefix + "_" + string(rune('a'+i))
	}
	return squad
}

func TestPositionAnchor(t *testing.T) {
	tests := []struct {
		position string
		wantX    int
		wantY    int
	}{
		{"goalkeeper", 6, 50},
		{"left_back", 21, 18},
		{"right_back", 21, 82},
		{"center_back", 17, 50},
		{"defensive_midfielder", 34, 50},
		{"left_midfielder", 40, 20},
		{"right_midfielder", 40, 80},
		{"left_wingback", 40, 20},
		{"right_wingback", 40, 80},
		{"left_winger", 40, 20},
		{"right_winger", 40, 80},
		{"central_midfielder", 45, 50},
		{"attacking_midfielder", 58, 50},
		{"second_striker", 66, 50},
		{"striker", 76, 50},
		{"center_forward", 76, 50},
		{"unknown_position", 44, 50},
		{"", 44, 50},
	}

	for _, tt := range tests {
		t.Run(tt.position, func(t *testing.T) {
			anchor := positionAnchor(tt.position)
			if anchor.X != tt.wantX || anchor.Y != tt.wantY {
				t.Errorf("positionAnchor(%q) = (%d, %d), want (%d, %d)",
					tt.position, anchor.X, anchor.Y, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestCanonicalPositionEdgeCases(t *testing.T) {
	tests := []struct {
		position string
		role     string
		want     string
	}{
		{"", "goalkeeper", "goalkeeper"},
		{"auto", "defender", "center_back"},
		{"goleiro", "goalkeeper", "goalkeeper"},
		{"keeper", "goalkeeper", "goalkeeper"},
		{"zagueiro", "defender", "center_back"},
		{"centerback", "defender", "center_back"},
		{"lateralesquerdo", "defender", "left_back"},
		{"lateraldireito", "defender", "right_back"},
		{"volante", "midfielder", "defensive_midfielder"},
		{"meia", "midfielder", "central_midfielder"},
		{"meiocampista", "midfielder", "central_midfielder"},
		{"meiaofensivo", "midfielder", "attacking_midfielder"},
		{"alaesquerda", "midfielder", "left_winger"},
		{"pontaesquerda", "midfielder", "left_winger"},
		{"aladireita", "midfielder", "right_winger"},
		{"pontadireita", "midfielder", "right_winger"},
		{"segundoatacante", "forward", "second_striker"},
		{"centroavante", "forward", "center_forward"},
		{"unknown", "midfielder", "central_midfielder"},
		{"GOLEIRO", "goalkeeper", "goalkeeper"},
	}

	for _, tt := range tests {
		t.Run(tt.position+"_"+tt.role, func(t *testing.T) {
			got := canonicalPosition(tt.position, tt.role)
			if got != tt.want {
				t.Errorf("canonicalPosition(%q, %q) = %q, want %q",
					tt.position, tt.role, got, tt.want)
			}
		})
	}
}

func TestDefaultTeamName(t *testing.T) {
	tests := []struct {
		name   string
		isHome bool
		want   string
	}{
		{"Team A", true, "Team A"},
		{"Team B", false, "Team B"},
		{"", true, "Casa"},
		{"  ", true, "Casa"},
		{"", false, "Fora"},
		{"  ", false, "Fora"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+boolString(tt.isHome), func(t *testing.T) {
			got := defaultTeamName(tt.name, tt.isHome)
			if got != tt.want {
				t.Errorf("defaultTeamName(%q, %v) = %q, want %q",
					tt.name, tt.isHome, got, tt.want)
			}
		})
	}
}

func boolString(b bool) string {
	if b {
		return "home"
	}
	return "away"
}

func TestClampStatValue(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{50, 50},
		{-10, 1},
		{0, 1},
		{1, 1},
		{99, 99},
		{100, 99},
		{150, 99},
	}

	for _, tt := range tests {
		got := clampStatValue(tt.input)
		if got != tt.want {
			t.Errorf("clampStatValue(%d) = %d, want %d",
				tt.input, got, tt.want)
		}
	}
}

func TestClampInt(t *testing.T) {
	tests := []struct {
		input int
		min   int
		max   int
		want  int
	}{
		{5, 0, 10, 5},
		{-5, 0, 10, 0},
		{15, 0, 10, 10},
		{0, 0, 10, 0},
		{10, 0, 10, 10},
	}

	for _, tt := range tests {
		got := clampInt(tt.input, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("clampInt(%d, %d, %d) = %d, want %d",
				tt.input, tt.min, tt.max, got, tt.want)
		}
	}
}

func TestSnapshotTeamName(t *testing.T) {
	tests := []struct {
		name     string
		snapshot DebugSnapshot
		isHome   bool
		want     string
	}{
		{
			name: "home with name",
			snapshot: DebugSnapshot{
				Field: FieldSnapshot{
					HomeName: "Home Team",
					AwayName: "Away Team",
				},
			},
			isHome: true,
			want:   "Home Team",
		},
		{
			name: "away with name",
			snapshot: DebugSnapshot{
				Field: FieldSnapshot{
					HomeName: "Home Team",
					AwayName: "Away Team",
				},
			},
			isHome: false,
			want:   "Away Team",
		},
		{
			name: "home without name",
			snapshot: DebugSnapshot{
				Field: FieldSnapshot{},
			},
			isHome: true,
			want:   "Casa",
		},
		{
			name: "away without name",
			snapshot: DebugSnapshot{
				Field: FieldSnapshot{},
			},
			isHome: false,
			want:   "Fora",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := snapshotTeamName(tt.snapshot, tt.isHome)
			if got != tt.want {
				t.Errorf("snapshotTeamName() = %q, want %q", got, tt.want)
			}
		})
	}
}