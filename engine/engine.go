package engine

import (
	"fmt"
	"manager/game/internal/domain/player"
	"manager/game/internal/domain/training"
	"manager/game/simulation"
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

func (e * Engine) FinishTraining(session training.Training) {
    //todo: Finalizar a traning session
}

func (e *Engine) PlayMatch() {
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