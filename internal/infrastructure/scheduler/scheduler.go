package scheduler

import (
	"context"
	"log"
	"manager/game/engine"
	"manager/game/internal/config"
	"manager/game/internal/domain/calendar"
	dbrepository "manager/game/internal/infrastructure/database/generated"
	trainingrepository "manager/game/internal/infrastructure/database/repository"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	engine  *engine.Engine
	queries *dbrepository.Queries
	config  config.Config
	clock   calendar.ServerClock
}

func New(engine *engine.Engine, queries *dbrepository.Queries, cfg config.Config) *Scheduler {
	tickDuration := time.Duration(cfg.CalendarTickSeconds) * time.Second
	if tickDuration <= 0 {
		tickDuration = time.Minute
	}

	return &Scheduler{
		engine:  engine,
		queries: queries,
		config:  cfg,
		clock:   calendar.NewServerClock(time.Local, tickDuration),
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	if s.config.SimulationTickCron != "" {
		s.runCron(ctx)
		return
	}

	ticker := time.NewTicker(time.Duration(s.config.SimulationTickSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.Tick(ctx)
		}
	}
}

func (s *Scheduler) runCron(ctx context.Context) {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	scheduler := cron.New(cron.WithParser(parser))

	_, err := scheduler.AddFunc(s.config.SimulationTickCron, func() {
		s.Tick(ctx)
	})
	if err != nil {
		log.Printf("scheduler cron parse error: %v", err)
		return
	}

	scheduler.Start()
	defer scheduler.Stop()

	<-ctx.Done()
}

func (s *Scheduler) Tick(ctx context.Context) {
	start := time.Now()
	now := s.clock.Now()
	serverTick := s.clock.TickAt(now)

	workerPoolSize := s.config.SimulationWorkerPoolSize
	if workerPoolSize <= 0 {
		workerPoolSize = s.config.SimulationMaxParallel
	}
	if workerPoolSize <= 0 {
		workerPoolSize = runtime.NumCPU()
	}
	batchSize := s.config.SimulationMatchBatchSize
	if batchSize <= 0 {
		batchSize = 128
	}
	queueSize := s.config.SimulationQueueSize
	if queueSize <= 0 {
		queueSize = workerPoolSize * 2
	}
	if queueSize < 1 {
		queueSize = 1
	}

	trainings := trainingrepository.FindPendingTrainings(now)
	matches, err := s.queries.ListInProgressMatches(ctx, int32(batchSize))
	if err != nil {
		log.Printf("scheduler: unable to list in-progress matches: %v", err)
		return
	}

	tasks := make([]func(), 0, len(trainings)+len(matches))

	for _, t := range trainings {
		training := t
		tasks = append(tasks, func() {
			s.engine.FinishTraining(training)
		})
	}

	for _, m := range matches {
		match := m
		tasks = append(tasks, func() {
			if err := s.processMatchTick(ctx, match); err != nil {
				log.Printf("scheduler: failed match %s tick %d: %v", match.ID, match.CurrentTick, err)
			}
		})
	}

	if len(matches) == 0 {
		tasks = append(tasks, func() {
			s.engine.ProcessNoMatchWindowTick(serverTick, now)
		})
	}

	s.runParallelTasks(tasks, workerPoolSize, queueSize)
	log.Printf("scheduler: tick finished in %s (server_tick=%d trainings=%d matches=%d workers=%d queue=%d)", time.Since(start), serverTick, len(trainings), len(matches), workerPoolSize, queueSize)
}

func (s *Scheduler) runParallelTasks(tasks []func(), workerPoolSize, queueSize int) {
	if len(tasks) == 0 {
		return
	}
	if workerPoolSize <= 0 {
		workerPoolSize = 1
	}
	if queueSize < 1 {
		queueSize = 1
	}

	jobs := make(chan func(), queueSize)
	var wg sync.WaitGroup

	for i := 0; i < workerPoolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range jobs {
				task()
			}
		}()
	}

	for _, task := range tasks {
		jobs <- task
	}
	close(jobs)

	wg.Wait()
}

func (s *Scheduler) processMatchTick(ctx context.Context, m dbrepository.Match) error {
	outcome := s.engine.PlayMatchTick(engine.PlayMatchTickInput{
		MatchID:     m.ID,
		CurrentTick: int(m.CurrentTick),
		Seed:        m.RandomSeed,
		HomeClubID:  m.HomeClubID,
		AwayClubID:  m.AwayClubID,
		HomeScore:   int(m.HomeScore),
		AwayScore:   int(m.AwayScore),
	})

	if _, err := s.queries.CreateMatchEvent(ctx, dbrepository.CreateMatchEventParams{
		MatchID:     m.ID,
		Tick:        int16(outcome.Tick),
		EventType:   outcome.EventType,
		Description: outcome.Description,
		HomeScore:   int32(outcome.HomeScore),
		AwayScore:   int32(outcome.AwayScore),
		Payload:     []byte("{}"),
	}); err != nil {
		return err
	}

	if outcome.IsFinished {
		winnerID, loserID, isDraw := resolveWinnerLoser(m.HomeClubID, m.AwayClubID, outcome.HomeScore, outcome.AwayScore)

		if _, err := s.queries.UpsertMatchResult(ctx, dbrepository.UpsertMatchResultParams{
			MatchID:       m.ID,
			HomeTeamScore: int32(outcome.HomeScore),
			AwayTeamScore: int32(outcome.AwayScore),
			WinnerClubID:  winnerID,
			LoserClubID:   loserID,
			IsDraw:        isDraw,
		}); err != nil {
			return err
		}

		if err := s.queries.UpdateMatchScoreAndTick(ctx, dbrepository.UpdateMatchScoreAndTickParams{
			ID:          m.ID,
			HomeScore:   int32(outcome.HomeScore),
			AwayScore:   int32(outcome.AwayScore),
			CurrentTick: int16(outcome.Tick),
		}); err != nil {
			return err
		}

		return s.queries.FinishMatch(ctx, m.ID)
	}

	return s.queries.UpdateMatchScoreAndTick(ctx, dbrepository.UpdateMatchScoreAndTickParams{
		ID:          m.ID,
		HomeScore:   int32(outcome.HomeScore),
		AwayScore:   int32(outcome.AwayScore),
		CurrentTick: int16(outcome.NextTick),
	})
}

func resolveWinnerLoser(homeClubID, awayClubID uuid.UUID, homeScore, awayScore int) (uuid.NullUUID, uuid.NullUUID, bool) {
	if homeScore == awayScore {
		return uuid.NullUUID{Valid: false}, uuid.NullUUID{Valid: false}, true
	}

	if homeScore > awayScore {
		return uuid.NullUUID{UUID: homeClubID, Valid: true}, uuid.NullUUID{UUID: awayClubID, Valid: true}, false
	}

	return uuid.NullUUID{UUID: awayClubID, Valid: true}, uuid.NullUUID{UUID: homeClubID, Valid: true}, false
}
