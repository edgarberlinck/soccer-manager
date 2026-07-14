package api

import (
	"context"
	"database/sql"
	"fmt"
	repository "manager/game/internal/infrastructure/database/generated"
	mathrand "math/rand"
	"time"

	"github.com/google/uuid"
)

var starterPositions = []string{
	"GK", "GK",
	"DF", "DF", "DF", "DF", "DF", "DF", "DF", "DF",
	"MF", "MF", "MF", "MF", "MF", "MF", "MF", "MF",
	"FW", "FW", "FW", "FW",
}

var starterFirstNames = []string{"Leo", "Bruno", "Caio", "Davi", "Enzo", "Theo", "Rafa", "Iago", "Luca", "Nico", "Joao", "Pedro", "Marcos", "Vitor", "Andre", "Diego", "Fabio", "Gustavo"}
var starterLastNames = []string{"Silva", "Santos", "Costa", "Oliveira", "Souza", "Rocha", "Almeida", "Pereira", "Melo", "Ribeiro", "Cardoso", "Freitas", "Barbosa", "Moura", "Teixeira"}

func bootstrapStarterSquad(ctx context.Context, queries *repository.Queries, clubID uuid.UUID) error {
	rng := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	contractStart := time.Now().UTC()
	contractEnd := contractStart.AddDate(3, 0, 0)

	for i, position := range starterPositions {
		tier := drawTier(rng)
		overall := baseOverallByTier(rng, tier)
		potential := minInt16(95, overall+int16(randomBetween(rng, 2, 11)))

		playerID := uuid.New()
		_, err := queries.CreateClubPlayer(ctx, repository.CreateClubPlayerParams{
			ID:            playerID,
			ClubID:        uuid.NullUUID{UUID: clubID, Valid: true},
			Name:          randomName(rng),
			Age:           int32(randomBetween(rng, 17, 33)),
			Position:      position,
			Overall:       overall,
			Potential:     potential,
			Tier:          tier,
			Pace:          clampAttribute(int16(overall + int16(randomBetween(rng, -10, 8)))),
			Passing:       clampAttribute(int16(overall + int16(randomBetween(rng, -9, 9)))),
			Shooting:      clampAttribute(int16(overall + int16(randomBetween(rng, -8, 10)))),
			Altura:        clampAttribute(int16(randomBetween(rng, 60, 95))),
			Peso:          clampAttribute(int16(randomBetween(rng, 55, 95))),
			Impulso:       clampAttribute(int16(overall + int16(randomBetween(rng, -7, 8)))),
			Explosao:      clampAttribute(int16(overall + int16(randomBetween(rng, -7, 8)))),
			Fisico:        clampAttribute(int16(overall + int16(randomBetween(rng, -8, 7)))),
			FisicalStatus: clampAttribute(int16(overall + int16(randomBetween(rng, -9, 7)))),
			Cabeceio:      clampAttribute(int16(overall + int16(randomBetween(rng, -10, 8)))),
			Cruzamento:    clampAttribute(int16(overall + int16(randomBetween(rng, -8, 9)))),
			Habilidade:    clampAttribute(int16(overall + int16(randomBetween(rng, -8, 10)))),
			Finalizacao:   clampAttribute(int16(overall + int16(randomBetween(rng, -8, 10)))),
			Dominio:       clampAttribute(int16(overall + int16(randomBetween(rng, -8, 10)))),
			Temperamento:  clampAttribute(int16(randomBetween(rng, 55, 95))),
		})
		if err != nil {
			return fmt.Errorf("create player %d: %w", i, err)
		}

		salaryCents := int64(overall) * int64(randomBetween(rng, 700, 1500))
		_, err = queries.CreatePlayerContract(ctx, repository.CreatePlayerContractParams{
			ID:                 uuid.New(),
			PlayerID:           playerID,
			SalaryCents:        salaryCents,
			ReleaseClauseCents: sql.NullInt64{Int64: salaryCents * 18, Valid: true},
			StartsAt:           contractStart,
			EndsAt:             contractEnd,
		})
		if err != nil {
			return fmt.Errorf("create contract %d: %w", i, err)
		}
	}

	return nil
}

func randomName(rng *mathrand.Rand) string {
	first := starterFirstNames[randomBetween(rng, 0, len(starterFirstNames)-1)]
	last := starterLastNames[randomBetween(rng, 0, len(starterLastNames)-1)]
	return first + " " + last
}

func randomBetween(rng *mathrand.Rand, min, max int) int {
	return rng.Intn(max-min+1) + min
}

func drawTier(rng *mathrand.Rand) string {
	roll := rng.Float64()
	if roll < 0.66 {
		return "Ruim"
	}
	if roll < 0.91 {
		return "Mediano"
	}
	return "Bom"
}

func baseOverallByTier(rng *mathrand.Rand, tier string) int16 {
	switch tier {
	case "Ruim":
		return int16(randomBetween(rng, 45, 61))
	case "Mediano":
		return int16(randomBetween(rng, 62, 76))
	default:
		return int16(randomBetween(rng, 77, 88))
	}
}

func minInt16(a, b int16) int16 {
	if a < b {
		return a
	}
	return b
}

func clampAttribute(v int16) int16 {
	if v < 1 {
		return 1
	}
	if v > 99 {
		return 99
	}
	return v
}
