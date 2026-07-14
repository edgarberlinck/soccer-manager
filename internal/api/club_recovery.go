package api

import (
	"context"
	"database/sql"
	"fmt"
	repository "manager/game/internal/infrastructure/database/generated"
	"strings"
	"time"

	"github.com/google/uuid"
)

func ensureUserHasClubAndSquad(ctx context.Context, queries *repository.Queries, userID uuid.UUID, email string) (repository.Club, bool, error) {
	clubs, err := queries.GetUserClubs(ctx, userID)
	if err != nil {
		return repository.Club{}, false, fmt.Errorf("list clubs: %w", err)
	}

	if len(clubs) > 0 {
		return clubs[0], false, nil
	}

	baseName := defaultClubNameFromEmail(email)
	club, err := createDefaultClubForUser(ctx, queries, userID, baseName)
	if err != nil {
		return repository.Club{}, false, fmt.Errorf("create default club: %w", err)
	}

	if err := bootstrapStarterSquad(ctx, queries, club.ID); err != nil {
		return repository.Club{}, false, fmt.Errorf("bootstrap squad: %w", err)
	}

	return club, true, nil
}

func createDefaultClubForUser(ctx context.Context, queries *repository.Queries, userID uuid.UUID, baseName string) (repository.Club, error) {
	candidate := baseName

	for attempt := 0; attempt < 8; attempt++ {
		club, err := queries.CreateClub(ctx, repository.CreateClubParams{
			ID:           uuid.New(),
			UserID:       userID,
			Name:         candidate,
			ShortName:    sql.NullString{String: candidate, Valid: true},
			Abbreviation: sql.NullString{String: defaultAbbreviation(candidate), Valid: true},
			Continent:    sql.NullString{String: "Europe", Valid: true},
			Country:      sql.NullString{String: "Portugal", Valid: true},
		})
		if err == nil {
			return club, nil
		}

		candidate = fmt.Sprintf("%s %d", baseName, time.Now().Unix()%10000+int64(attempt))
	}

	return repository.Club{}, fmt.Errorf("unable to create unique default club name")
}

func defaultClubNameFromEmail(email string) string {
	local := strings.TrimSpace(strings.Split(email, "@")[0])
	if local == "" {
		return "Meu Clube"
	}

	local = strings.ReplaceAll(local, ".", " ")
	local = strings.ReplaceAll(local, "_", " ")
	local = strings.ReplaceAll(local, "-", " ")
	local = strings.TrimSpace(local)
	if local == "" {
		return "Meu Clube"
	}

	parts := strings.Fields(local)
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		runes := []rune(strings.ToLower(parts[i]))
		runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
		parts[i] = string(runes)
	}

	return strings.Join(parts, " ") + " FC"
}
