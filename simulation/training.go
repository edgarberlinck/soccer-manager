package simulation

import (
	"manager/game/internal/domain/player"
	"manager/game/internal/domain/training"
	"time"
)

func StartTraining(player player.Player, session training.TrainingSession) training.Training {
	return training.Training{
		Player: player,
		StartedAt: time.Now(),
		EndsAt: time.Now().Add(session.Duration),
		Session: session,
		Status: 0,
	}
}