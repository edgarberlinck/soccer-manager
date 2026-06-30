package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestAuthQueries(t *testing.T) {
	queries, mock, cleanup := newMockQueries(t)
	defer cleanup()

	id := uuid.New()
	now := time.Now()
	verifyToken := sql.NullString{String: "token-1", Valid: true}
	nullTime := sql.NullTime{}

	userColumns := []string{"id", "username", "password_hash", "active", "email_verified_at", "verification_token", "verification_token_expires_at", "created_at", "updated_at"}

	mock.ExpectQuery(regexp.QuoteMeta(createUser)).
		WithArgs(id, "email@test.com", "hash", true, verifyToken, nullTime).
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(id, "email@test.com", "hash", true, nil, "token-1", nil, now, now))

	created, err := queries.CreateUser(context.Background(), CreateUserParams{
		ID:                         id,
		Username:                   "email@test.com",
		PasswordHash:               "hash",
		Active:                     true,
		VerificationToken:          verifyToken,
		VerificationTokenExpiresAt: nullTime,
	})
	if err != nil || created.ID != id {
		t.Fatalf("create user failed: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(getUserByEmail)).
		WithArgs("email@test.com").
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(id, "email@test.com", "hash", true, nil, "token-1", nil, now, now))

	_, err = queries.GetUserByEmail(context.Background(), "email@test.com")
	if err != nil {
		t.Fatalf("get user by email failed: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(getUserByID)).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(id, "email@test.com", "hash", true, nil, "token-1", nil, now, now))

	_, err = queries.GetUserByID(context.Background(), id)
	if err != nil {
		t.Fatalf("get user by id failed: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(verifyUserByToken)).
		WithArgs(verifyToken).
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(id, "email@test.com", "hash", true, now, nil, nil, now, now))

	_, err = queries.VerifyUserByToken(context.Background(), verifyToken)
	if err != nil {
		t.Fatalf("verify user failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestClubQueriesAndManyErrorBranches(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		queries, mock, cleanup := newMockQueries(t)
		defer cleanup()

		now := time.Now()
		clubID := uuid.New()
		userID := uuid.New()
		clubColumns := []string{"id", "user_id", "name", "short_name", "abbreviation", "continent", "country", "created_at", "updated_at"}

		mock.ExpectQuery(regexp.QuoteMeta(createClub)).
			WithArgs(clubID, userID, "FC Test", sql.NullString{}, sql.NullString{}, sql.NullString{}, sql.NullString{}).
			WillReturnRows(sqlmock.NewRows(clubColumns).AddRow(clubID, userID, "FC Test", nil, nil, nil, nil, now, now))

		_, err := queries.CreateClub(context.Background(), CreateClubParams{ID: clubID, UserID: userID, Name: "FC Test"})
		if err != nil {
			t.Fatalf("create club failed: %v", err)
		}

		mock.ExpectQuery(regexp.QuoteMeta(getClubByName)).
			WithArgs("FC Test").
			WillReturnRows(sqlmock.NewRows(clubColumns).AddRow(clubID, userID, "FC Test", nil, nil, nil, nil, now, now))

		_, err = queries.GetClubByName(context.Background(), "FC Test")
		if err != nil {
			t.Fatalf("get club failed: %v", err)
		}

		mock.ExpectQuery(regexp.QuoteMeta(getUserClubs)).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows(clubColumns).AddRow(clubID, userID, "FC Test", nil, nil, nil, nil, now, now))

		clubs, err := queries.GetUserClubs(context.Background(), userID)
		if err != nil {
			t.Fatalf("get user clubs failed: %v", err)
		}
		if len(clubs) != 1 {
			t.Fatalf("expected 1 club, got %d", len(clubs))
		}
	})

	t.Run("query error", func(t *testing.T) {
		queries, mock, cleanup := newMockQueries(t)
		defer cleanup()
		userID := uuid.New()

		mock.ExpectQuery(regexp.QuoteMeta(getUserClubs)).
			WithArgs(userID).
			WillReturnError(errors.New("query failed"))

		_, err := queries.GetUserClubs(context.Background(), userID)
		if err == nil {
			t.Fatal("expected query error")
		}
	})

	t.Run("scan error", func(t *testing.T) {
		queries, mock, cleanup := newMockQueries(t)
		defer cleanup()

		userID := uuid.New()
		rows := sqlmock.NewRows([]string{"id"}).AddRow("not-uuid")
		mock.ExpectQuery(regexp.QuoteMeta(getUserClubs)).WithArgs(userID).WillReturnRows(rows)

		_, err := queries.GetUserClubs(context.Background(), userID)
		if err == nil {
			t.Fatal("expected scan error")
		}
	})

	t.Run("rows error", func(t *testing.T) {
		queries, mock, cleanup := newMockQueries(t)
		defer cleanup()

		userID := uuid.New()
		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "short_name", "abbreviation", "continent", "country", "created_at", "updated_at"}).
			AddRow(uuid.New(), userID, "FC", nil, nil, nil, nil, now, now).
			RowError(0, errors.New("row err"))
		mock.ExpectQuery(regexp.QuoteMeta(getUserClubs)).WithArgs(userID).WillReturnRows(rows)

		_, err := queries.GetUserClubs(context.Background(), userID)
		if err == nil {
			t.Fatal("expected row error")
		}
	})
}

func TestPlayerQueries(t *testing.T) {
	queries, mock, cleanup := newMockQueries(t)
	defer cleanup()

	id := uuid.New()
	playerCols := []string{"id", "name", "age", "pace", "passing", "shooting"}

	mock.ExpectQuery(regexp.QuoteMeta(createPlayer)).
		WithArgs(id, "Ronaldo", int32(30), int16(90), int16(80), int16(85)).
		WillReturnRows(sqlmock.NewRows(playerCols).AddRow(id, "Ronaldo", int32(30), int16(90), int16(80), int16(85)))
	_, err := queries.CreatePlayer(context.Background(), CreatePlayerParams{ID: id, Name: "Ronaldo", Age: 30, Pace: 90, Passing: 80, Shooting: 85})
	if err != nil {
		t.Fatalf("create player failed: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(getPlayer)).WithArgs(id).
		WillReturnRows(sqlmock.NewRows(playerCols).AddRow(id, "Ronaldo", int32(30), int16(90), int16(80), int16(85)))
	_, err = queries.GetPlayer(context.Background(), id)
	if err != nil {
		t.Fatalf("get player failed: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(increasePlayerAge)).WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := queries.IncreasePlayerAge(context.Background(), id); err != nil {
		t.Fatalf("increase age failed: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(updatePlayer)).
		WithArgs(id, "Ronaldo", int32(31), int16(91), int16(81), int16(86)).
		WillReturnRows(sqlmock.NewRows(playerCols).AddRow(id, "Ronaldo", int32(31), int16(91), int16(81), int16(86)))
	_, err = queries.UpdatePlayer(context.Background(), UpdatePlayerParams{ID: id, Name: "Ronaldo", Age: 31, Pace: 91, Passing: 81, Shooting: 86})
	if err != nil {
		t.Fatalf("update player failed: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(updatePlayerAttributes)).
		WithArgs(id, int16(92), int16(82), int16(87)).
		WillReturnRows(sqlmock.NewRows(playerCols).AddRow(id, "Ronaldo", int32(31), int16(92), int16(82), int16(87)))
	_, err = queries.UpdatePlayerAttributes(context.Background(), UpdatePlayerAttributesParams{ID: id, Pace: 92, Passing: 82, Shooting: 87})
	if err != nil {
		t.Fatalf("update attributes failed: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(findPlayersReadyToRetire)).WithArgs(int32(30)).
		WillReturnRows(sqlmock.NewRows(playerCols).AddRow(id, "Ronaldo", int32(30), int16(90), int16(80), int16(85)))
	list, err := queries.FindPlayersReadyToRetire(context.Background(), 30)
	if err != nil || len(list) != 1 {
		t.Fatalf("find retire failed: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(listPlayers)).
		WillReturnRows(sqlmock.NewRows(playerCols).AddRow(id, "Ronaldo", int32(30), int16(90), int16(80), int16(85)))
	list, err = queries.ListPlayers(context.Background())
	if err != nil || len(list) != 1 {
		t.Fatalf("list players failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPlayerManyErrorBranches(t *testing.T) {
	queries, mock, cleanup := newMockQueries(t)
	defer cleanup()

	t.Run("find query error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(findPlayersReadyToRetire)).WithArgs(int32(40)).WillReturnError(errors.New("query failed"))
		_, err := queries.FindPlayersReadyToRetire(context.Background(), 40)
		if err == nil {
			t.Fatal("expected query error")
		}
	})

	t.Run("list query error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(listPlayers)).WillReturnError(errors.New("query failed"))
		_, err := queries.ListPlayers(context.Background())
		if err == nil {
			t.Fatal("expected query error")
		}
	})
}

func TestUserMetaQueries(t *testing.T) {
	queries, mock, cleanup := newMockQueries(t)
	defer cleanup()

	userID := uuid.New()
	now := time.Now()
	social := json.RawMessage(`{"x":"y"}`)
	metadata := json.RawMessage(`{"z":1}`)
	cols := []string{"user_id", "full_name", "country", "social_links", "metadata", "created_at", "updated_at"}

	mock.ExpectQuery(regexp.QuoteMeta(getUserMetaByUserID)).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(userID, "User", "BR", social, metadata, now, now))
	_, err := queries.GetUserMetaByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("get user meta failed: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(upsertUserMeta)).
		WithArgs(userID, sql.NullString{String: "User", Valid: true}, sql.NullString{String: "BR", Valid: true}, social, metadata).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(userID, "User", "BR", social, metadata, now, now))
	_, err = queries.UpsertUserMeta(context.Background(), UpsertUserMetaParams{
		UserID:      userID,
		FullName:    sql.NullString{String: "User", Valid: true},
		Country:     sql.NullString{String: "BR", Valid: true},
		SocialLinks: social,
		Metadata:    metadata,
	})
	if err != nil {
		t.Fatalf("upsert user meta failed: %v", err)
	}
}
