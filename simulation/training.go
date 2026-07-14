package simulation

import (
	"manager/game/internal/domain/player"
	"manager/game/internal/domain/training"
	"time"
)

func StartTraining(player player.Player, session training.TrainingSession) training.Training {
	session.Duration = session.ResolvedDuration()
	startedAt := time.Now()
	endsAt := startedAt.Add(session.Duration)

	return training.Training{
		Player:    player,
		StartedAt: startedAt,
		EndsAt:    endsAt,
		Session:   session,
		Status:    0,
	}
}

func FinishTraining(current training.Training, now time.Time, randomFactor float64) training.Training {
	gain := training.ComputeTrainingGain(current.Session, current.Player, randomFactor)
	fatigue := training.ComputeFatigueCost(current.Session)

	updatedPlayer := current.Player
	updatedPlayer.Attributes = current.Session.Type.ApplyGain(updatedPlayer.Attributes, gain)
	updatedPlayer.Attributes.FisicalStatus = training.ClampPhysicalStatus(updatedPlayer.Attributes.FisicalStatus - fatigue)

	current.Player = updatedPlayer
	current.Status = 1
	if now.After(current.EndsAt) {
		current.EndsAt = now
	}
	current.Summary = training.Summary{
		TrainingSession: current.Session,
		Score:           float64(gain),
	}

	return current
}