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
	if result.Session.Duration != session.ResolvedDuration() {
		t.Fatalf("expected resolved duration %s, got %s", session.ResolvedDuration(), result.Session.Duration)
	}
	if result.EndsAt.Before(result.StartedAt) {
		t.Fatal("expected training end after start")
	}
}

func TestFinishTrainingMarksSessionFinishedAndAppliesGain(t *testing.T) {
	athlete := player.Player{
		Name: "Tester",
		Age:  21,
		Attributes: player.Attributes{
			Passing:       50,
			FisicalStatus: 90,
		},
	}
	session := training.TrainingSession{Intensity: training.Intense, Type: training.Passing}
	started := StartTraining(athlete, session)
	finished := FinishTraining(started, started.EndsAt, 1.0)

	if finished.Status != 1 {
		t.Fatalf("expected finished status, got %d", finished.Status)
	}
	if finished.Player.Attributes.Passing <= athlete.Attributes.Passing {
		t.Fatalf("expected passing to improve, got before=%d after=%d", athlete.Attributes.Passing, finished.Player.Attributes.Passing)
	}
}

func TestFinishTrainingReducesPhysicalStatusWithoutLeavingBounds(t *testing.T) {
	athlete := player.Player{
		Name: "Tester",
		Age:  28,
		Attributes: player.Attributes{
			Pace:          40,
			FisicalStatus: 4,
		},
	}
	session := training.TrainingSession{Intensity: training.Intense, Type: training.Speed}
	started := StartTraining(athlete, session)
	finished := FinishTraining(started, started.EndsAt, 1.0)

	if finished.Player.Attributes.FisicalStatus < 0 {
		t.Fatalf("expected physical status to stay >= 0, got %d", finished.Player.Attributes.FisicalStatus)
	}
	if finished.Player.Attributes.FisicalStatus >= athlete.Attributes.FisicalStatus {
		t.Fatalf("expected physical status to decrease, got before=%d after=%d", athlete.Attributes.FisicalStatus, finished.Player.Attributes.FisicalStatus)
	}
	if finished.Player.Attributes.FisicalStatus > 100 {
		t.Fatalf("expected physical status to stay <= 100, got %d", finished.Player.Attributes.FisicalStatus)
	}
	if finished.Player.Attributes.Pace <= athlete.Attributes.Pace {
		t.Fatalf("expected speed training to improve pace, got before=%d after=%d", athlete.Attributes.Pace, finished.Player.Attributes.Pace)
	}
}
