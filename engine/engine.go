package engine

import (
	"fmt"
	"manager/game/internal/domain/club"
	"manager/game/internal/domain/player"
	"manager/game/internal/domain/training"
	"manager/game/simulation"

	"github.com/google/uuid"
)

type Engine struct {
	// futuras dependencias
}

func New() *Engine {
	return &Engine{
		// aqui eu posso inicializar as dependencias
	}
}

func (e *Engine) StartTraining(player player.Player, session training.TrainingSession) {
	s := simulation.StartTraining(player, session)
	fmt.Println(s)
}

func (e *Engine) FinishTraining(session training.Training) {
	// todo: Finalizar a traning session
}

type PlayMatchTickInput struct {
	MatchID     uuid.UUID
	CurrentTick int
	Seed        int64
	HomeClubID  uuid.UUID
	AwayClubID  uuid.UUID
	HomeScore   int
	AwayScore   int
}

func (e *Engine) PlayMatchTick(input PlayMatchTickInput) simulation.TickOutcome {
	return simulation.PlayMatchTick(simulation.PlayMatchTickInput{
		MatchID:     input.MatchID,
		CurrentTick: input.CurrentTick,
		Seed:        input.Seed,
		HomeClubID:  input.HomeClubID,
		AwayClubID:  input.AwayClubID,
		HomeScore:   input.HomeScore,
		AwayScore:   input.AwayScore,
	})
}

func (e *Engine) PlayMatch() {
	_ = simulation.PlayMatch(club.Club{}, club.Club{})
	fmt.Println("playing a match")
}

func (e *Engine) TrainPlayer() {
	fmt.Println("training a player")
}

func (e *Engine) StartSeason() {
	fmt.Println("starting season")
}

func (e *Engine) FinishSeason() {
	fmt.Println("finishing the season")
}

func (e *Engine) PaySalaries() {
	fmt.Println("paying salaries")
}

func (e *Engine) RetirePlayer() {
	fmt.Println("retiring player")
}
