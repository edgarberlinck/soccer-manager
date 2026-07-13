package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"manager/game/internal/domain/calendar"
	"manager/game/internal/domain/player"
	"manager/game/simulation"
	"os"
	"path/filepath"
	"sort"
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

type reportCalendar struct {
	ServerNow      string              `json:"server_now"`
	DayStart       string              `json:"day_start"`
	DayEnd         string              `json:"day_end"`
	TickNow        int                 `json:"tick_now"`
	DayStartTick   int                 `json:"day_start_tick"`
	DayEndTick     int                 `json:"day_end_tick"`
	Sporting       []calendar.CalendarEntry `json:"sporting"`
	Administrative []calendar.CalendarEntry `json:"administrative"`
	AgendaNow      calendar.TickAgenda `json:"agenda_now"`
	AgendaMatchPlan calendar.SimulationBatchPlan `json:"agenda_match_plan"`
}

type simulationReport struct {
	Seed       int64                     `json:"seed"`
	GeneratedAt string                   `json:"generated_at"`
	MatchID    uuid.UUID                 `json:"match_id"`
	HomeTeam   string                    `json:"home_team"`
	AwayTeam   string                    `json:"away_team"`
	HomeScore  int                       `json:"home_score"`
	AwayScore  int                       `json:"away_score"`
	Calendar   reportCalendar            `json:"calendar"`
	PerformanceSummary reportPerformanceSummary `json:"performance_summary"`
	Snapshots  []simulation.DebugSnapshot `json:"snapshots"`
}

type reportPerformanceSummary struct {
	Home reportTeamLeaders `json:"home"`
	Away reportTeamLeaders `json:"away"`
}

type reportTeamLeaders struct {
	Team           string             `json:"team"`
	Movement       []reportStatLeader `json:"movement"`
	Touches        []reportStatLeader `json:"touches"`
	CorrectTouches []reportStatLeader `json:"correct_touches"`
	LongPasses     []reportStatLeader `json:"long_passes"`
	ShotsOnGoal    []reportStatLeader `json:"shots_on_goal"`
	Fouls          []reportStatLeader `json:"fouls"`
}

type reportStatLeader struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Value      int    `json:"value"`
}

func main() {
	seedFlag := flag.Int64("seed", 0, "0 usa seed aleatoria; qualquer outro valor reproduz a simulacao")
	outFlag := flag.String("out", "./tmp/simulation-output.json", "arquivo JSON de saída com os dados gerados")
	tickSecondsFlag := flag.Int("calendar-tick-seconds", 60, "duração de cada tick do calendário em segundos")
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

	report, err := createSimulationReport(time.Now(), seed, matchID, home, away, snapshots, *tickSecondsFlag)
	if err != nil {
		log.Fatal(err)
	}
	if err := writeReport(*outFlag, report); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("SIMULACAO DEBUG MANUAL - sem efeito no fluxo normal da API | seed=%d\n", seed)
	fmt.Printf("Dados gerados em: %s\n", *outFlag)
	for index, snapshot := range snapshots {
		if index > 0 {
			fmt.Println()
			fmt.Println(strings.Repeat("=", 72))
		}
		fmt.Println(simulation.RenderDebugSnapshot(snapshot))
	}

	printPerformanceSummary(report)
}

func createSimulationReport(now time.Time, seed int64, matchID uuid.UUID, home, away simulation.DebugTeam, snapshots []simulation.DebugSnapshot, tickSeconds int) (simulationReport, error) {
	if len(snapshots) == 0 {
		return simulationReport{}, errors.New("simulation generated no snapshots")
	}

	clock := calendar.NewServerClock(time.Local, time.Duration(tickSeconds)*time.Second)
	dayStart, dayEnd := clock.DayBounds(now)
	dayStartTick, dayEndTick := clock.TickRangeForDay(now)
	tickNow := clock.TickAt(now)

	matchStart := tickNow + 5
	if matchStart > dayEndTick {
		matchStart = dayEndTick
	}
	matchEnd := matchStart + len(snapshots) - 1
	if matchEnd > dayEndTick {
		matchEnd = dayEndTick
	}

	seasonCalendar, err := calendar.BuildSeasonCalendar(calendar.BuildSeasonCalendarInput{
		ClubID:          home.ClubID,
		SeasonStartTick: dayStartTick,
		SeasonEndTick:   dayEndTick,
		Matches: []calendar.MatchSlot{
			{
				MatchID:        matchID,
				Kind:           calendar.EntryChampionshipMatch,
				Title:          "championship_match",
				StartTick:      matchStart,
				EndTick:        matchEnd,
				OpponentClubID: away.ClubID,
			},
		},
		TransferWindows: []calendar.TransferWindow{
			{
				WindowID:  uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("transfer:%s", home.ClubID))),
				Title:     "transfer_window",
				StartTick: dayStartTick,
				EndTick:   dayEndTick,
			},
		},
	})
	if err != nil {
		return simulationReport{}, err
	}

	agendaNow := seasonCalendar.AgendaAt(tickNow)
	agendaPlan := calendar.PlanMatchSimulation(agendaNow.ActiveMatches, 2, 2)
	final := snapshots[len(snapshots)-1]

	return simulationReport{
		Seed:        seed,
		GeneratedAt: now.Format(time.RFC3339),
		MatchID:     matchID,
		HomeTeam:    home.Name,
		AwayTeam:    away.Name,
		HomeScore:   final.Outcome.HomeScore,
		AwayScore:   final.Outcome.AwayScore,
		Calendar: reportCalendar{
			ServerNow:       now.In(time.Local).Format(time.RFC3339),
			DayStart:        dayStart.Format(time.RFC3339),
			DayEnd:          dayEnd.Format(time.RFC3339),
			TickNow:         tickNow,
			DayStartTick:    dayStartTick,
			DayEndTick:      dayEndTick,
			Sporting:        seasonCalendar.Sporting,
			Administrative:  seasonCalendar.Administrative,
			AgendaNow:       agendaNow,
			AgendaMatchPlan: agendaPlan,
		},
		PerformanceSummary: buildPerformanceSummary(final, home.Name, away.Name),
		Snapshots: snapshots,
	}, nil
}

func buildPerformanceSummary(final simulation.DebugSnapshot, homeName, awayName string) reportPerformanceSummary {
	return reportPerformanceSummary{
		Home: buildTeamLeaders(homeName, final.Field.HomeSquad),
		Away: buildTeamLeaders(awayName, final.Field.AwaySquad),
	}
}

func buildTeamLeaders(teamName string, members []simulation.SquadMemberSnapshot) reportTeamLeaders {
	return reportTeamLeaders{
		Team:           teamName,
		Movement:       topLeaders(members, func(s simulation.PlayerMatchStats) int { return s.Movement }),
		Touches:        topLeaders(members, func(s simulation.PlayerMatchStats) int { return s.Touches }),
		CorrectTouches: topLeaders(members, func(s simulation.PlayerMatchStats) int { return s.CorrectTouches }),
		LongPasses:     topLeaders(members, func(s simulation.PlayerMatchStats) int { return s.LongPasses }),
		ShotsOnGoal:    topLeaders(members, func(s simulation.PlayerMatchStats) int { return s.ShotsOnGoal }),
		Fouls:          topLeaders(members, func(s simulation.PlayerMatchStats) int { return s.Fouls }),
	}
}

func topLeaders(members []simulation.SquadMemberSnapshot, valueFn func(simulation.PlayerMatchStats) int) []reportStatLeader {
	leaders := make([]reportStatLeader, 0, len(members))
	for _, member := range members {
		leaders = append(leaders, reportStatLeader{
			PlayerID:   member.ID,
			PlayerName: member.Name,
			Value:      valueFn(member.Stats),
		})
	}

	sort.SliceStable(leaders, func(i, j int) bool {
		if leaders[i].Value == leaders[j].Value {
			return leaders[i].PlayerName < leaders[j].PlayerName
		}
		return leaders[i].Value > leaders[j].Value
	})

	if len(leaders) > 3 {
		leaders = leaders[:3]
	}
	return leaders
}

func printPerformanceSummary(report simulationReport) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 72))
	fmt.Println("RESUMO DE PERFORMANCE (TOP 3)")
	printTeamLeaders(report.PerformanceSummary.Home)
	printTeamLeaders(report.PerformanceSummary.Away)
}

func printTeamLeaders(team reportTeamLeaders) {
	fmt.Printf("\nTime: %s\n", team.Team)
	printLeaderLine("Movimentacao", team.Movement)
	printLeaderLine("Toques", team.Touches)
	printLeaderLine("Toques corretos", team.CorrectTouches)
	printLeaderLine("Lancamentos", team.LongPasses)
	printLeaderLine("Chutes no gol", team.ShotsOnGoal)
	printLeaderLine("Faltas", team.Fouls)
}

func printLeaderLine(label string, leaders []reportStatLeader) {
	entries := make([]string, 0, len(leaders))
	for _, leader := range leaders {
		entries = append(entries, fmt.Sprintf("%s (%d)", leader.PlayerName, leader.Value))
	}
	fmt.Printf("- %s: %s\n", label, strings.Join(entries, ", "))
}

func writeReport(path string, report simulationReport) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("report path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
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
	if len(raw.Players) < 11 {
		return simulation.DebugTeam{}, fmt.Errorf("arquivo %s precisa ter pelo menos 11 jogadores, recebeu %d", path, len(raw.Players))
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