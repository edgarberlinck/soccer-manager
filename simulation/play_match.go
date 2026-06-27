package simulation

import (
	"manager/game/internal/domain/club"
	"manager/game/internal/domain/match"
)

func PlayMatch(home, away club.Club) match.Result {
	// Aqui a mágica vai acontecer. Tenho que estimar o resultado da partida
	
	return match.Result{
		HomeTeamScore: 1,
		AwayTeamScore: 1,
	}
}