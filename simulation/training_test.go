package simulation

import (
	"testing"
	"time"

	"manager/game/internal/domain/player"
	"manager/game/internal/domain/training"
)

func TestStartTraining(t *testing.T) {
	session := training.TrainingSession{Duration: 90 * time.Minute, Intensity: training.Medium, Type: training.Passing}
	athlete := player.Player{Name: "Tester"}

	result := StartTraining(athlete, session)

	if result.Player.Name != "Tester" {
		t.Fatalf("expected player Tester, got %s", result.Player.Name)
	}
	if result.Session.Duration != 90*time.Minute {
		t.Fatalf("expected 90 minutes duration, got %s", result.Session.Duration)
	}
	if result.EndsAt.Before(result.StartedAt) {
		t.Fatal("expected training end after start")
	}
}
