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

func TestManyQueriesCloseAndRowsErrors(t *testing.T) {
	now := time.Now()

	t.Run("GetUserClubs close error", func(t *testing.T) {
		queries, mock, cleanup := newMockQueries(t)
		defer cleanup()

		userID := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "short_name", "abbreviation", "continent", "country", "created_at", "updated_at"}).
			AddRow(uuid.New(), userID, "FC", nil, nil, nil, nil, now, now).
			CloseError(errors.New("close error"))
		mock.ExpectQuery(regexp.QuoteMeta(getUserClubs)).WithArgs(userID).WillReturnRows(rows)

		_, err := queries.GetUserClubs(context.Background(), userID)
		if err == nil {
			t.Fatal("expected close error")
		}
	})

	t.Run("ListInProgressMatches close error", func(t *testing.T) {
		queries, mock, cleanup := newMockQueries(t)
		defer cleanup()

		rows := sqlmock.NewRows([]string{"id", "home_club_id", "away_club_id", "championship_id", "status", "current_tick", "random_seed", "home_score", "away_score", "finished_at", "created_at", "updated_at"}).
			AddRow(uuid.New(), uuid.New(), uuid.New(), nil, "in_progress", int16(1), int64(1), int32(0), int32(0), nil, now, now).
			CloseError(errors.New("close error"))
		mock.ExpectQuery(regexp.QuoteMeta(listInProgressMatches)).WithArgs(int32(20)).WillReturnRows(rows)

		_, err := queries.ListInProgressMatches(context.Background(), 20)
		if err == nil {
			t.Fatal("expected close error")
		}
	})

	t.Run("FindPlayersReadyToRetire close error", func(t *testing.T) {
		queries, mock, cleanup := newMockQueries(t)
		defer cleanup()

		rows := sqlmock.NewRows([]string{"id", "name", "age", "pace", "passing", "shooting"}).
			AddRow(uuid.New(), "A", int32(40), int16(1), int16(1), int16(1)).
			CloseError(errors.New("close error"))
		mock.ExpectQuery(regexp.QuoteMeta(findPlayersReadyToRetire)).WithArgs(int32(40)).WillReturnRows(rows)

		_, err := queries.FindPlayersReadyToRetire(context.Background(), 40)
		if err == nil {
			t.Fatal("expected close error")
		}
	})

	t.Run("ListPlayers close error", func(t *testing.T) {
		queries, mock, cleanup := newMockQueries(t)
		defer cleanup()

		rows := sqlmock.NewRows([]string{"id", "name", "age", "pace", "passing", "shooting"}).
			AddRow(uuid.New(), "A", int32(20), int16(1), int16(1), int16(1)).
			CloseError(errors.New("close error"))
		mock.ExpectQuery(regexp.QuoteMeta(listPlayers)).WillReturnRows(rows)

		_, err := queries.ListPlayers(context.Background())
		if err == nil {
			t.Fatal("expected close error")
		}
	})
}
