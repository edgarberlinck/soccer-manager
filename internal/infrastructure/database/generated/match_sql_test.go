package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func newMockQueries(t *testing.T) (*Queries, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	cleanup := func() {
		_ = db.Close()
	}

	return New(db), mock, cleanup
}

func TestCreateMatch(t *testing.T) {
	queries, mock, cleanup := newMockQueries(t)
	defer cleanup()

	id := uuid.New()
	home := uuid.New()
	away := uuid.New()
	now := time.Now()

	mock.ExpectQuery("INSERT INTO \"match\"").
		WithArgs(id, home, away, uuid.NullUUID{Valid: false}, int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "home_club_id", "away_club_id", "championship_id", "status", "current_tick", "random_seed", "home_score", "away_score", "finished_at", "created_at", "updated_at"}).
			AddRow(id, home, away, nil, "in_progress", int16(1), int64(11), int32(0), int32(0), nil, now, now))

	result, err := queries.CreateMatch(context.Background(), CreateMatchParams{
		ID:         id,
		HomeClubID: home,
		AwayClubID: away,
		RandomSeed: 11,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != id {
		t.Fatalf("expected match id %s, got %s", id, result.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestListInProgressMatches(t *testing.T) {
	queries, mock, cleanup := newMockQueries(t)
	defer cleanup()

	id := uuid.New()
	home := uuid.New()
	away := uuid.New()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(listInProgressMatches)).
		WithArgs(int32(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "home_club_id", "away_club_id", "championship_id", "status", "current_tick", "random_seed", "home_score", "away_score", "finished_at", "created_at", "updated_at"}).
			AddRow(id, home, away, nil, "in_progress", int16(2), int64(7), int32(1), int32(0), nil, now, now))

	matches, err := queries.ListInProgressMatches(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected one match, got %d", len(matches))
	}
	if matches[0].ID != id {
		t.Fatalf("expected id %s, got %s", id, matches[0].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestListInProgressMatchesQueryError(t *testing.T) {
	queries, mock, cleanup := newMockQueries(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(listInProgressMatches)).
		WithArgs(int32(10)).
		WillReturnError(errors.New("query failed"))

	_, err := queries.ListInProgressMatches(context.Background(), 10)
	if err == nil {
		t.Fatal("expected query error")
	}
}

func TestCreateMatchEvent(t *testing.T) {
	queries, mock, cleanup := newMockQueries(t)
	defer cleanup()

	matchID := uuid.New()
	now := time.Now()
	payload := []byte("{}")

	mock.ExpectQuery(regexp.QuoteMeta(createMatchEvent)).
		WithArgs(matchID, int16(3), "short_pass", "Tick 3: Passe", int32(0), int32(0), payload).
		WillReturnRows(sqlmock.NewRows([]string{"id", "match_id", "tick", "event_type", "description", "home_score", "away_score", "payload", "created_at"}).
			AddRow(int64(1), matchID, int16(3), "short_pass", "Tick 3: Passe", int32(0), int32(0), payload, now))

	event, err := queries.CreateMatchEvent(context.Background(), CreateMatchEventParams{
		MatchID:     matchID,
		Tick:        3,
		EventType:   "short_pass",
		Description: "Tick 3: Passe",
		HomeScore:   0,
		AwayScore:   0,
		Payload:     payload,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.MatchID != matchID {
		t.Fatalf("expected match id %s, got %s", matchID, event.MatchID)
	}
}

func TestUpdateAndFinishMatch(t *testing.T) {
	queries, mock, cleanup := newMockQueries(t)
	defer cleanup()

	matchID := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta(updateMatchScoreAndTick)).
		WithArgs(matchID, int32(2), int32(1), int16(33)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := queries.UpdateMatchScoreAndTick(context.Background(), UpdateMatchScoreAndTickParams{
		ID:          matchID,
		HomeScore:   2,
		AwayScore:   1,
		CurrentTick: 33,
	})
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(finishMatch)).
		WithArgs(matchID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := queries.FinishMatch(context.Background(), matchID); err != nil {
		t.Fatalf("unexpected finish error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUpsertMatchResult(t *testing.T) {
	queries, mock, cleanup := newMockQueries(t)
	defer cleanup()

	matchID := uuid.New()
	homeID := uuid.New()
	awayID := uuid.New()
	now := time.Now()

	winner := uuid.NullUUID{UUID: homeID, Valid: true}
	loser := uuid.NullUUID{UUID: awayID, Valid: true}

	mock.ExpectQuery(regexp.QuoteMeta(upsertMatchResult)).
		WithArgs(matchID, int32(3), int32(1), winner, loser, false).
		WillReturnRows(sqlmock.NewRows([]string{"match_id", "home_team_score", "away_team_score", "winner_club_id", "loser_club_id", "is_draw", "created_at"}).
			AddRow(matchID, int32(3), int32(1), homeID, awayID, false, now))

	result, err := queries.UpsertMatchResult(context.Background(), UpsertMatchResultParams{
		MatchID:       matchID,
		HomeTeamScore: 3,
		AwayTeamScore: 1,
		WinnerClubID:  winner,
		LoserClubID:   loser,
		IsDraw:        false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MatchID != matchID {
		t.Fatalf("expected match id %s, got %s", matchID, result.MatchID)
	}
}

func TestWithTx(t *testing.T) {
	queries, _, cleanup := newMockQueries(t)
	defer cleanup()

	txQueries := queries.WithTx(nil)
	if txQueries == nil {
		t.Fatal("expected non-nil tx queries")
	}
}
