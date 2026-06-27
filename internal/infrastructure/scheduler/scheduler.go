package scheduler

import (
	"context"
	"database/sql"
	"manager/game/engine"
	"manager/game/internal/infrastructure/database/repository"
	"time"
)

type Scheduler struct {
	engine *engine.Engine
	db *sql.DB
}

func New(engine *engine.Engine, db *sql.DB) *Scheduler {
	return &Scheduler{
		engine: engine,
		db: db,
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker((time.Minute))

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): return
		case <-ticker.C: s.Tick()
		}
	}
}

func (s *Scheduler) Tick() {

	trainings := repository.FindPendingTrainings(time.Now())

	for _, t := range trainings {

		s.engine.FinishTraining(t)

	}
}