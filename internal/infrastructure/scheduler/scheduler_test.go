package scheduler

import (
	"context"
	"errors"
	"regexp"
	"sync/atomic"
	"testing"
	"time"

	"manager/game/engine"
	"manager/game/internal/config"
	dbrepository "manager/game/internal/infrastructure/database/generated"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func newSchedulerWithMock(t *testing.T, cfg config.Config) (*Scheduler, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	queries := dbrepository.New(db)
	s := New(engine.New(), queries, cfg)

	cleanup := func() {
		_ = db.Close()
	}

	return s, mock, cleanup
}

func TestRunWithCancelledContext(t *testing.T) {
	s, _, cleanup := newSchedulerWithMock(t, config.Config{SimulationTickSeconds: 1})
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s.Run(ctx)
}

func TestRunCronInvalidExpression(t *testing.T) {
	s, _, cleanup := newSchedulerWithMock(t, config.Config{SimulationTickCron: "invalid cron"})
	defer cleanup()

	s.Run(context.Background())
}

func TestRunCronValidAndCancelled(t *testing.T) {
	s, _, cleanup := newSchedulerWithMock(t, config.Config{SimulationTickCron: "*/1 * * * * *"})
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)
		s.Run(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("scheduler did not stop after cron context cancellation")
	}
}

func TestRunTickerTriggersTick(t *testing.T) {
	s, mock, cleanup := newSchedulerWithMock(t, config.Config{SimulationTickSeconds: 1})
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`-- name: ListInProgressMatches :many
SELECT id, home_club_id, away_club_id, championship_id, status, current_tick, random_seed, home_score, away_score, finished_at, created_at, updated_at
FROM "match"
WHERE status = 'in_progress'
ORDER BY updated_at ASC
LIMIT $1
`)).WithArgs(int32(128)).WillReturnError(errors.New("db down"))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)
		s.Run(ctx)
	}()

	time.Sleep(1100 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("scheduler did not stop after ticker context cancellation")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestTickListInProgressMatchesError(t *testing.T) {
	s, mock, cleanup := newSchedulerWithMock(t, config.Config{})
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`-- name: ListInProgressMatches :many
SELECT id, home_club_id, away_club_id, championship_id, status, current_tick, random_seed, home_score, away_score, finished_at, created_at, updated_at
FROM "match"
WHERE status = 'in_progress'
ORDER BY updated_at ASC
LIMIT $1
`)).WithArgs(int32(128)).WillReturnError(errors.New("db down"))

	s.Tick(context.Background())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRunParallelTasks(t *testing.T) {
	s, _, cleanup := newSchedulerWithMock(t, config.Config{})
	defer cleanup()

	var count int32
	tasks := []func(){
		func() { atomic.AddInt32(&count, 1) },
		func() { atomic.AddInt32(&count, 1) },
		func() { atomic.AddInt32(&count, 1) },
	}

	s.runParallelTasks(tasks, 2)

	if got := atomic.LoadInt32(&count); got != 3 {
		t.Fatalf("expected all tasks executed, got %d", got)
	}

	s.runParallelTasks(nil, 2)
}

func TestProcessMatchTickInProgress(t *testing.T) {
	s, mock, cleanup := newSchedulerWithMock(t, config.Config{})
	defer cleanup()

	matchID := uuid.New()
	homeID := uuid.New()
	awayID := uuid.New()

	m := dbrepository.Match{
		ID:          matchID,
		HomeClubID:  homeID,
		AwayClubID:  awayID,
		CurrentTick: 2,
		RandomSeed:  77,
		HomeScore:   0,
		AwayScore:   0,
	}

	outcome := s.engine.PlayMatchTick(engine.PlayMatchTickInput{
		MatchID:     m.ID,
		CurrentTick: int(m.CurrentTick),
		Seed:        m.RandomSeed,
		HomeClubID:  m.HomeClubID,
		AwayClubID:  m.AwayClubID,
		HomeScore:   int(m.HomeScore),
		AwayScore:   int(m.AwayScore),
	})

	mock.ExpectQuery("INSERT INTO match_events").
		WithArgs(m.ID, int16(outcome.Tick), outcome.EventType, outcome.Description, int32(outcome.HomeScore), int32(outcome.AwayScore), []byte("{}")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "match_id", "tick", "event_type", "description", "home_score", "away_score", "payload", "created_at"}).
			AddRow(int64(1), m.ID, int16(outcome.Tick), outcome.EventType, outcome.Description, int32(outcome.HomeScore), int32(outcome.AwayScore), []byte("{}"), time.Now()))

	mock.ExpectExec("UPDATE \"match\"").
		WithArgs(m.ID, int32(outcome.HomeScore), int32(outcome.AwayScore), int16(outcome.NextTick)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := s.processMatchTick(context.Background(), m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestProcessMatchTickFinished(t *testing.T) {
	s, mock, cleanup := newSchedulerWithMock(t, config.Config{})
	defer cleanup()

	matchID := uuid.New()
	homeID := uuid.New()
	awayID := uuid.New()

	m := dbrepository.Match{
		ID:          matchID,
		HomeClubID:  homeID,
		AwayClubID:  awayID,
		CurrentTick: 90,
		RandomSeed:  77,
		HomeScore:   10,
		AwayScore:   0,
	}

	outcome := s.engine.PlayMatchTick(engine.PlayMatchTickInput{
		MatchID:     m.ID,
		CurrentTick: int(m.CurrentTick),
		Seed:        m.RandomSeed,
		HomeClubID:  m.HomeClubID,
		AwayClubID:  m.AwayClubID,
		HomeScore:   int(m.HomeScore),
		AwayScore:   int(m.AwayScore),
	})

	mock.ExpectQuery("INSERT INTO match_events").
		WithArgs(m.ID, int16(outcome.Tick), outcome.EventType, outcome.Description, int32(outcome.HomeScore), int32(outcome.AwayScore), []byte("{}")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "match_id", "tick", "event_type", "description", "home_score", "away_score", "payload", "created_at"}).
			AddRow(int64(1), m.ID, int16(outcome.Tick), outcome.EventType, outcome.Description, int32(outcome.HomeScore), int32(outcome.AwayScore), []byte("{}"), time.Now()))

	mock.ExpectQuery("INSERT INTO match_results").
		WithArgs(m.ID, int32(outcome.HomeScore), int32(outcome.AwayScore), sqlmock.AnyArg(), sqlmock.AnyArg(), false).
		WillReturnRows(sqlmock.NewRows([]string{"match_id", "home_team_score", "away_team_score", "winner_club_id", "loser_club_id", "is_draw", "created_at"}).
			AddRow(m.ID, int32(outcome.HomeScore), int32(outcome.AwayScore), homeID, awayID, false, time.Now()))

	mock.ExpectExec("UPDATE \"match\"").
		WithArgs(m.ID, int32(outcome.HomeScore), int32(outcome.AwayScore), int16(outcome.Tick)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE \"match\"").
		WithArgs(m.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := s.processMatchTick(context.Background(), m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestProcessMatchTickErrorPaths(t *testing.T) {
	t.Run("create event", func(t *testing.T) {
		s, mock, cleanup := newSchedulerWithMock(t, config.Config{})
		defer cleanup()

		m := dbrepository.Match{ID: uuid.New(), HomeClubID: uuid.New(), AwayClubID: uuid.New(), CurrentTick: 2}
		mock.ExpectQuery("INSERT INTO match_events").WillReturnError(errors.New("insert failed"))

		if err := s.processMatchTick(context.Background(), m); err == nil {
			t.Fatal("expected error from CreateMatchEvent")
		}
	})

	t.Run("upsert result", func(t *testing.T) {
		s, mock, cleanup := newSchedulerWithMock(t, config.Config{})
		defer cleanup()

		m := dbrepository.Match{ID: uuid.New(), HomeClubID: uuid.New(), AwayClubID: uuid.New(), CurrentTick: 90, HomeScore: 10, AwayScore: 0}
		outcome := s.engine.PlayMatchTick(engine.PlayMatchTickInput{MatchID: m.ID, CurrentTick: 90, HomeClubID: m.HomeClubID, AwayClubID: m.AwayClubID, HomeScore: 10, AwayScore: 0})

		mock.ExpectQuery("INSERT INTO match_events").
			WillReturnRows(sqlmock.NewRows([]string{"id", "match_id", "tick", "event_type", "description", "home_score", "away_score", "payload", "created_at"}).
				AddRow(int64(1), m.ID, int16(outcome.Tick), outcome.EventType, outcome.Description, int32(outcome.HomeScore), int32(outcome.AwayScore), []byte("{}"), time.Now()))
		mock.ExpectQuery("INSERT INTO match_results").WillReturnError(errors.New("upsert failed"))

		if err := s.processMatchTick(context.Background(), m); err == nil {
			t.Fatal("expected error from UpsertMatchResult")
		}
	})

	t.Run("update score and tick", func(t *testing.T) {
		s, mock, cleanup := newSchedulerWithMock(t, config.Config{})
		defer cleanup()

		m := dbrepository.Match{ID: uuid.New(), HomeClubID: uuid.New(), AwayClubID: uuid.New(), CurrentTick: 2}
		outcome := s.engine.PlayMatchTick(engine.PlayMatchTickInput{MatchID: m.ID, CurrentTick: 2, HomeClubID: m.HomeClubID, AwayClubID: m.AwayClubID})

		mock.ExpectQuery("INSERT INTO match_events").
			WillReturnRows(sqlmock.NewRows([]string{"id", "match_id", "tick", "event_type", "description", "home_score", "away_score", "payload", "created_at"}).
				AddRow(int64(1), m.ID, int16(outcome.Tick), outcome.EventType, outcome.Description, int32(outcome.HomeScore), int32(outcome.AwayScore), []byte("{}"), time.Now()))
		mock.ExpectExec("UPDATE \"match\"").WillReturnError(errors.New("update failed"))

		if err := s.processMatchTick(context.Background(), m); err == nil {
			t.Fatal("expected error from UpdateMatchScoreAndTick")
		}
	})

	t.Run("finish match", func(t *testing.T) {
		s, mock, cleanup := newSchedulerWithMock(t, config.Config{})
		defer cleanup()

		m := dbrepository.Match{ID: uuid.New(), HomeClubID: uuid.New(), AwayClubID: uuid.New(), CurrentTick: 90, HomeScore: 10, AwayScore: 0}
		outcome := s.engine.PlayMatchTick(engine.PlayMatchTickInput{MatchID: m.ID, CurrentTick: 90, HomeClubID: m.HomeClubID, AwayClubID: m.AwayClubID, HomeScore: 10, AwayScore: 0})

		mock.ExpectQuery("INSERT INTO match_events").
			WillReturnRows(sqlmock.NewRows([]string{"id", "match_id", "tick", "event_type", "description", "home_score", "away_score", "payload", "created_at"}).
				AddRow(int64(1), m.ID, int16(outcome.Tick), outcome.EventType, outcome.Description, int32(outcome.HomeScore), int32(outcome.AwayScore), []byte("{}"), time.Now()))
		mock.ExpectQuery("INSERT INTO match_results").
			WillReturnRows(sqlmock.NewRows([]string{"match_id", "home_team_score", "away_team_score", "winner_club_id", "loser_club_id", "is_draw", "created_at"}).
				AddRow(m.ID, int32(outcome.HomeScore), int32(outcome.AwayScore), m.HomeClubID, m.AwayClubID, false, time.Now()))
		mock.ExpectExec("UPDATE \"match\"").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("UPDATE \"match\"").WillReturnError(errors.New("finish failed"))

		if err := s.processMatchTick(context.Background(), m); err == nil {
			t.Fatal("expected error from FinishMatch")
		}
	})
}

func TestResolveWinnerLoser(t *testing.T) {
	home := uuid.New()
	away := uuid.New()

	winner, loser, draw := resolveWinnerLoser(home, away, 2, 2)
	if !draw || winner.Valid || loser.Valid {
		t.Fatal("expected draw with no winner/loser")
	}

	winner, loser, draw = resolveWinnerLoser(home, away, 3, 1)
	if draw || !winner.Valid || winner.UUID != home || !loser.Valid || loser.UUID != away {
		t.Fatal("expected home winner")
	}

	winner, loser, draw = resolveWinnerLoser(home, away, 0, 1)
	if draw || !winner.Valid || winner.UUID != away || !loser.Valid || loser.UUID != home {
		t.Fatal("expected away winner")
	}
}

func TestNew(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	queries := dbrepository.New(db)
	s := New(engine.New(), queries, config.Config{})
	if s == nil {
		t.Fatal("expected scheduler instance")
	}
	if s.queries == nil {
		t.Fatal("expected queries in scheduler")
	}
}
