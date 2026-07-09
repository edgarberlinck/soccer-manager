package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"manager/game/internal/domain/player"
	"manager/game/simulation"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type squadFile struct {
	TeamName string            `json:"team_name"`
	ClubID   string            `json:"club_id"`
	Players  []squadFilePlayer `json:"players"`
}

type squadFilePlayer struct {
	Name       string            `json:"name"`
	Role       string            `json:"role"`
	Position   string            `json:"position"`
	Attributes player.Attributes `json:"attributes"`
}

func main() {
	seedFlag := flag.Int64("seed", 0, "0 usa seed aleatoria; qualquer outro valor reproduz a simulacao")
	flag.Parse()

	if flag.NArg() != 2 {
		log.Fatalf("uso: go run ./cmd/simulate-debug [-seed N] <time_casa.json> <time_fora.json>")
	}

	seed := *seedFlag
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	home, err := loadDebugTeam(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	away, err := loadDebugTeam(flag.Arg(1))
	if err != nil {
		log.Fatal(err)
	}

	matchID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(home.Name+"::"+away.Name))
	snapshots := simulation.SimulateDebugMatch(simulation.DebugMatchConfig{
		MatchID: matchID,
		Seed:    seed,
		Home:    home,
		Away:    away,
	})

	fmt.Printf("SIMULACAO DEBUG MANUAL - sem efeito no fluxo normal da API | seed=%d\n", seed)
	for index, snapshot := range snapshots {
		if index > 0 {
			fmt.Println()
			fmt.Println(strings.Repeat("=", 72))
		}
		fmt.Println(simulation.RenderDebugSnapshot(snapshot))
	}
}

func loadDebugTeam(path string) (simulation.DebugTeam, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return simulation.DebugTeam{}, fmt.Errorf("falha ao ler %s: %w", path, err)
	}

	var raw squadFile
	if err := json.Unmarshal(content, &raw); err != nil {
		return simulation.DebugTeam{}, fmt.Errorf("json invalido em %s: %w", path, err)
	}
	if len(raw.Players) != 11 {
		return simulation.DebugTeam{}, fmt.Errorf("arquivo %s precisa ter exatamente 11 jogadores, recebeu %d", path, len(raw.Players))
	}

	clubID, err := resolveClubID(raw, path)
	if err != nil {
		return simulation.DebugTeam{}, err
	}

	players := make([]simulation.TacticalPlayer, 0, len(raw.Players))
	for index, rawPlayer := range raw.Players {
		if strings.TrimSpace(rawPlayer.Name) == "" {
			return simulation.DebugTeam{}, fmt.Errorf("jogador %d em %s esta sem nome", index+1, path)
		}
		if strings.TrimSpace(rawPlayer.Role) == "" {
			return simulation.DebugTeam{}, fmt.Errorf("jogador %s em %s esta sem role", rawPlayer.Name, path)
		}

		players = append(players, simulation.TacticalPlayer{
			Name:       rawPlayer.Name,
			Role:       rawPlayer.Role,
			Position:   rawPlayer.Position,
			Attributes: rawPlayer.Attributes,
		})
	}

	teamName := strings.TrimSpace(raw.TeamName)
	if teamName == "" {
		teamName = strings.TrimSuffix(filepath.Base(path), ".json")
	}

	return simulation.DebugTeam{
		Name:    teamName,
		ClubID:  clubID,
		Players: players,
	}, nil
}

func resolveClubID(raw squadFile, path string) (uuid.UUID, error) {
	if strings.TrimSpace(raw.ClubID) != "" {
		clubID, err := uuid.Parse(raw.ClubID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("club_id invalido em %s: %w", path, err)
		}
		return clubID, nil
	}

	base := strings.TrimSpace(raw.TeamName)
	if base == "" {
		return uuid.Nil, errors.New("team_name ou club_id precisam ser informados")
	}
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(base)), nil
}