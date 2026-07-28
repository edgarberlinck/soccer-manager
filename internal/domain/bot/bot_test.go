package bot

import (
	"manager/game/internal/domain/player"
	"manager/game/internal/domain/training"
	"testing"
)

func TestBot_DecideTraining(t *testing.T) {
	tests := []struct {
		name     string
		strategy Strategy
		player   player.Player
	}{
		{
			name:     "Conservador com jogador cansado",
			strategy: Conservador,
			player: player.Player{
				Attributes: player.Attributes{
					FisicalStatus: 30,
				},
			},
		},
		{
			name:     "Agressivo com jogador descansado",
			strategy: Agressivo,
			player: player.Player{
				Attributes: player.Attributes{
					FisicalStatus: 90,
				},
			},
		},
		{
			name:     "Equilibrado com jogador médio",
			strategy: Equilibrado,
			player: player.Player{
				Attributes: player.Attributes{
					FisicalStatus: 60,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := Bot{
				ID:       "test-bot",
				Username: "test",
				Strategy: tt.strategy,
				ClubID:   "test-club",
			}

			session := bot.DecideTraining(tt.player)

			if session.Type < training.Finishing || session.Type > training.Goalkeeping {
				t.Errorf("Tipo de treino inválido: %v", session.Type)
			}

			if session.Intensity < training.Soft || session.Intensity > training.Intense {
				t.Errorf("Intensidade inválida: %v", session.Intensity)
			}
		})
	}
}

func TestBot_ShouldRest(t *testing.T) {
	tests := []struct {
		name           string
		strategy       Strategy
		physicalStatus int
		wantRest       bool
	}{
		{"Conservador - status baixo", Conservador, 50, true},
		{"Conservador - status ok", Conservador, 70, false},
		{"Equilibrado - status baixo", Equilibrado, 35, true},
		{"Equilibrado - status ok", Equilibrado, 50, false},
		{"Agressivo - status muito baixo", Agressivo, 20, true},
		{"Agressivo - status baixo ok", Agressivo, 30, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := Bot{Strategy: tt.strategy}
			player := player.Player{
				Attributes: player.Attributes{
					FisicalStatus: tt.physicalStatus,
				},
			}

			got := bot.ShouldRest(player)
			if got != tt.wantRest {
				t.Errorf("ShouldRest() = %v, want %v", got, tt.wantRest)
			}
		})
	}
}

func TestBot_SelectBestPlayers(t *testing.T) {
	bot := Bot{Strategy: Equilibrado}

	players := []player.Player{
		{
			Id:   "1",
			Name: "Weak Player",
			Attributes: player.Attributes{
				Pace:          50,
				Passing:       50,
				Shooting:      50,
				Fisico:        50,
				FisicalStatus: 50,
			},
		},
		{
			Id:   "2",
			Name: "Strong Player",
			Attributes: player.Attributes{
				Pace:          80,
				Passing:       80,
				Shooting:      80,
				Fisico:        80,
				FisicalStatus: 80,
			},
		},
		{
			Id:   "3",
			Name: "Medium Player",
			Attributes: player.Attributes{
				Pace:          65,
				Passing:       65,
				Shooting:      65,
				Fisico:        65,
				FisicalStatus: 65,
			},
		},
	}

	selected := bot.SelectBestPlayers(players, 2)

	if len(selected) != 2 {
		t.Errorf("Expected 2 players, got %d", len(selected))
	}

	// O melhor jogador deve ser o primeiro
	if selected[0].Id != "2" {
		t.Errorf("Expected strongest player first, got %s", selected[0].Name)
	}

	// O segundo melhor deve ser o segundo
	if selected[1].Id != "3" {
		t.Errorf("Expected medium player second, got %s", selected[1].Name)
	}
}

func TestCalculateOverall(t *testing.T) {
	player := player.Player{
		Attributes: player.Attributes{
			Pace:          70,
			Passing:       80,
			Shooting:      75,
			Fisico:        85,
			FisicalStatus: 90,
			Cabeceio:      65,
			Cruzamento:    70,
			Habilidade:    75,
			Finalizacao:   80,
			Dominio:       85,
		},
	}

	overall := calculateOverall(player)
	expected := 77.5 // (70+80+75+85+90+65+70+75+80+85) / 10

	if overall != expected {
		t.Errorf("Expected overall %.1f, got %.1f", expected, overall)
	}
}
