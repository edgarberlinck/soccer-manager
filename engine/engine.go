package engine

import (
	"fmt"
	"manager/game/internal/domain/calendar"
	"manager/game/internal/domain/club"
	"manager/game/internal/domain/player"
	"manager/game/internal/domain/training"
	"manager/game/simulation"
	"time"

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

func (e *Engine) StartTraining(player player.Player, session training.TrainingSession) training.Training {
	return simulation.StartTraining(player, session)
}

func (e *Engine) FinishTraining(session training.Training) training.Training {
	return simulation.FinishTraining(session, time.Now(), 1.0)
}

func (e *Engine) BuildSeasonCalendar(input calendar.BuildSeasonCalendarInput) (calendar.SeasonCalendar, error) {
	return calendar.BuildSeasonCalendar(input)
}

func (e *Engine) CalendarAgendaAt(calendarState calendar.SeasonCalendar, tick int) calendar.TickAgenda {
	return calendarState.AgendaAt(tick)
}

func (e *Engine) PlanCalendarMatchSimulation(entries []calendar.CalendarEntry, maxParallel, maxMatchesPerBatch int) calendar.SimulationBatchPlan {
	return calendar.PlanMatchSimulation(entries, maxParallel, maxMatchesPerBatch)
}

func (e *Engine) ProcessNoMatchWindowTick(serverTick int, now time.Time) {
	// TODO: encaixar regras de treino automático, scouting e mercado de transferências.
	fmt.Printf("idle-window tick=%d at=%s\n", serverTick, now.Format(time.RFC3339))
}

type PlayMatchTickInput struct {
	MatchID     uuid.UUID
	CurrentTick int
	Seed        int64
	HomeClubID  uuid.UUID
	AwayClubID  uuid.UUID
	HomeScore   int
	AwayScore   int
	PossessionTeam string
	BallZone string
	HomePlayers []simulation.TacticalPlayer
	AwayPlayers []simulation.TacticalPlayer
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
		PossessionTeam: input.PossessionTeam,
		BallZone: input.BallZone,
		HomePlayers: input.HomePlayers,
		AwayPlayers: input.AwayPlayers,
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
