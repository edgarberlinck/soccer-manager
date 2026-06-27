package repository

import (
	"manager/game/internal/domain/player"
	"manager/game/internal/domain/training"
	"time"
)

func FindPendingTrainings(t time.Time) []training.Training {
	return []training.Training{
		{
			Player: player.Player{
				Name: "Ronaldo",
			},
			StartedAt: time.Now().Add(-2 * time.Hour),
			EndsAt:    time.Now().Add(-10 * time.Minute),
			Status:    0,
		},	
	}
}