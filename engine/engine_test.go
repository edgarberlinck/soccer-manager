package engine

import (
	"manager/game/internal/domain/player"
	"manager/game/internal/domain/training"
	"testing"
	"time"
)

func TestEngineStartTrainingReturnsTraining(t *testing.T) {
	eng := New()
	athlete := player.Player{Name: "Tester", Attributes: player.Attributes{FisicalStatus: 90}}
	session := training.TrainingSession{Type: training.Passing, Intensity: training.Medium}

	result := eng.StartTraining(athlete, session)

	if result.Player.Name != athlete.Name {
		t.Fatalf("expected player %s, got %s", athlete.Name, result.Player.Name)
	}
	if result.Status != 0 {
		t.Fatalf("expected in-progress status, got %d", result.Status)
	}
}

func TestEngineFinishTrainingReturnsFinishedTraining(t *testing.T) {
	eng := New()
	athlete := player.Player{
		Name: "Tester",
		Age:  22,
		Attributes: player.Attributes{
			Passing:       40,
			FisicalStatus: 85,
		},
	}
	session := training.TrainingSession{Type: training.Passing, Intensity: training.Intense}
	started := eng.StartTraining(athlete, session)
	started.EndsAt = time.Now().Add(-time.Minute)

	finished := eng.FinishTraining(started)

	if finished.Status != 1 {
		t.Fatalf("expected finished status, got %d", finished.Status)
	}
	if finished.Player.Attributes.Passing <= athlete.Attributes.Passing {
		t.Fatalf("expected passing to improve, got before=%d after=%d", athlete.Attributes.Passing, finished.Player.Attributes.Passing)
	}
	if finished.Player.Attributes.FisicalStatus >= athlete.Attributes.FisicalStatus {
		t.Fatalf("expected physical status to be reduced, got before=%d after=%d", athlete.Attributes.FisicalStatus, finished.Player.Attributes.FisicalStatus)
	}
	if finished.Summary.Score <= 0 {
		t.Fatalf("expected positive summary score, got %f", finished.Summary.Score)
	}
}