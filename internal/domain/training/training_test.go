package training

import (
	"manager/game/internal/domain/player"
	"testing"
	"time"
)

func TestResolvedDurationVariesByIntensity(t *testing.T) {
	base := TrainingSession{Type: Passing, Intensity: Medium}.ResolvedDuration()
	light := TrainingSession{Type: Passing, Intensity: Soft}.ResolvedDuration()
	intense := TrainingSession{Type: Passing, Intensity: Intense}.ResolvedDuration()

	if !(light < base && base < intense) {
		t.Fatalf("expected Soft < Medium < Intense durations, got %s %s %s", light, base, intense)
	}
}

func TestComputeTrainingGainFallsWithAge(t *testing.T) {
	session := TrainingSession{Type: Passing, Intensity: Medium}
	younger := player.Player{Age: 20, Attributes: player.Attributes{FisicalStatus: 90}}
	older := player.Player{Age: 34, Attributes: player.Attributes{FisicalStatus: 90}}

	youngGain := ComputeTrainingGain(session, younger, 1.0)
	oldGain := ComputeTrainingGain(session, older, 1.0)

	if youngGain <= oldGain {
		t.Fatalf("expected younger player to gain more, got young=%d old=%d", youngGain, oldGain)
	}
}

func TestComputeTrainingGainFallsWithPhysicalStatus(t *testing.T) {
	session := TrainingSession{Type: Passing, Intensity: Medium}
	fit := player.Player{Age: 24, Attributes: player.Attributes{FisicalStatus: 90}}
	tired := player.Player{Age: 24, Attributes: player.Attributes{FisicalStatus: 35}}

	fitGain := ComputeTrainingGain(session, fit, 1.0)
	tiredGain := ComputeTrainingGain(session, tired, 1.0)

	if fitGain <= tiredGain {
		t.Fatalf("expected fitter player to gain more, got fit=%d tired=%d", fitGain, tiredGain)
	}
}

func TestComputeFatigueCostVariesByIntensity(t *testing.T) {
	light := ComputeFatigueCost(TrainingSession{Type: Passing, Intensity: Soft})
	medium := ComputeFatigueCost(TrainingSession{Type: Passing, Intensity: Medium})
	intense := ComputeFatigueCost(TrainingSession{Type: Passing, Intensity: Intense})

	if !(light < medium && medium < intense) {
		t.Fatalf("expected Soft < Medium < Intense fatigue, got %d %d %d", light, medium, intense)
	}
}

func TestResolvedDurationUsesTimeDuration(t *testing.T) {
	resolved := TrainingSession{Type: Speed, Intensity: Medium}.ResolvedDuration()
	if resolved < 30*time.Minute {
		t.Fatalf("expected resolved duration in minutes, got %s", resolved)
	}
}