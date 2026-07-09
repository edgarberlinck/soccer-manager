package simulation

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"strings"

	"github.com/google/uuid"
)

const (
	debugFieldWidth  = 41
	debugFieldHeight = 17

	ansiReset = "\033[0m"
	ansiBlue  = "\033[34m"
	ansiRed   = "\033[31m"
	ansiWhite = "\033[97m"
)

type DebugTeam struct {
	Name    string
	ClubID  uuid.UUID
	Players []TacticalPlayer
}

type DebugMatchConfig struct {
	MatchID uuid.UUID
	Seed    int64
	Home    DebugTeam
	Away    DebugTeam
}

type FieldPoint struct {
	X int
	Y int
}

type DebugActor struct {
	Name     string
	Role     string
	Position string
	Point    FieldPoint
}

type FieldSnapshot struct {
	Width      int
	Height     int
	HomeName   string
	AwayName   string
	Ball       FieldPoint
	Referee    FieldPoint
	Linesmen   [2]FieldPoint
	Home       []DebugActor
	Away       []DebugActor
	Possession string
	BallZone   string
	HomeMode   string
	AwayMode   string
}

type DebugSnapshot struct {
	Outcome TickOutcome
	Field   FieldSnapshot
}

func SimulateDebugMatch(cfg DebugMatchConfig) []DebugSnapshot {
	homePlayers := ensureTeamPlayers(cfg.Home.Players, cfg.Home.ClubID)
	awayPlayers := ensureTeamPlayers(cfg.Away.Players, cfg.Away.ClubID)

	possession := ""
	ballZone := zoneMiddle
	homeScore := 0
	awayScore := 0

	snapshots := make([]DebugSnapshot, 0, regulationTicks)

	for tick := 1; tick <= regulationTicks; tick++ {
		outcome := PlayMatchTick(PlayMatchTickInput{
			MatchID:        cfg.MatchID,
			CurrentTick:    tick,
			Seed:           cfg.Seed,
			HomeClubID:     cfg.Home.ClubID,
			AwayClubID:     cfg.Away.ClubID,
			HomeScore:      homeScore,
			AwayScore:      awayScore,
			PossessionTeam: possession,
			BallZone:       ballZone,
			EnvironmentNoise: true,
			HomePlayers:    homePlayers,
			AwayPlayers:    awayPlayers,
		})

		homeScore = outcome.HomeScore
		awayScore = outcome.AwayScore

		statePossession, stateZone := resolveDebugState(outcome, possession, ballZone)
		statePossession = nextPossession(statePossession, outcome.EventType)
		stateZone = advanceBallZone(stateZone, statePossession, outcome.EventType)
		homeMode, awayMode := resolveModes(statePossession, stateZone)
		field := buildFieldSnapshot(cfg, tick, statePossession, stateZone, homeMode, awayMode, homePlayers, awayPlayers)

		if outcome.PossessionTeam == "" {
			outcome.PossessionTeam = statePossession
		}
		if outcome.BallZone == "" {
			outcome.BallZone = stateZone
		}
		if outcome.HomeMode == "" {
			outcome.HomeMode = homeMode
		}
		if outcome.AwayMode == "" {
			outcome.AwayMode = awayMode
		}

		snapshots = append(snapshots, DebugSnapshot{
			Outcome: outcome,
			Field:   field,
		})

		possession = statePossession
		ballZone = stateZone
	}

	return snapshots
}

func RenderDebugSnapshot(snapshot DebugSnapshot) string {
	grid := make([][]string, debugFieldHeight)
	for y := 0; y < debugFieldHeight; y++ {
		grid[y] = make([]string, debugFieldWidth)
		for x := 0; x < debugFieldWidth; x++ {
			grid[y][x] = baseFieldCell(x, y)
		}
	}

	place := func(point FieldPoint, token string) {
		if point.Y < 0 || point.Y >= debugFieldHeight || point.X < 0 || point.X >= debugFieldWidth {
			return
		}
		grid[point.Y][point.X] = token
	}

	for _, actor := range snapshot.Field.Home {
		place(actor.Point, ansiBlue+"●"+ansiReset)
	}
	for _, actor := range snapshot.Field.Away {
		place(actor.Point, ansiRed+"●"+ansiReset)
	}
	place(snapshot.Field.Linesmen[0], "B")
	place(snapshot.Field.Linesmen[1], "B")
	place(snapshot.Field.Referee, "J")
	place(snapshot.Field.Ball, ansiWhite+"●"+ansiReset)

	var lines []string
	lines = append(lines, fmt.Sprintf(
		"Tick %02d | %s %d x %d %s | evento=%s | posse=%s | zona=%s | modos=%s/%s",
		snapshot.Outcome.Tick,
		snapshotTeamName(snapshot, true),
		snapshot.Outcome.HomeScore,
		snapshot.Outcome.AwayScore,
		snapshotTeamName(snapshot, false),
		snapshot.Outcome.EventType,
		snapshot.Field.Possession,
		snapshot.Field.BallZone,
		snapshot.Field.HomeMode,
		snapshot.Field.AwayMode,
	))
	lines = append(lines, snapshot.Outcome.Description)
	lines = append(lines, "Legenda: "+ansiBlue+"●"+ansiReset+" casa | "+ansiRed+"●"+ansiReset+" fora | "+ansiWhite+"●"+ansiReset+" bola | J juiz | B bandeirinha")

	for _, row := range grid {
		lines = append(lines, strings.Join(row, ""))
	}

	return strings.Join(lines, "\n")
}

func resolveDebugState(outcome TickOutcome, previousPossession, previousZone string) (string, string) {
	possession := outcome.PossessionTeam
	zone := outcome.BallZone

	if possession == "" {
		switch outcome.EventType {
		case "kickoff_home", "goal_home":
			possession = teamHome
		case "kickoff_away", "goal_away":
			possession = teamAway
		default:
			possession = previousPossession
		}
	}
	if possession == "" {
		possession = teamHome
	}

	if zone == "" {
		switch outcome.EventType {
		case "goal_home", "goal_away", "kickoff_home", "kickoff_away", "halftime", "fulltime":
			zone = zoneMiddle
		default:
			zone = previousZone
		}
	}
	if zone == "" {
		zone = zoneMiddle
	}

	return possession, zone
}

func advanceBallZone(currentZone, possession, eventType string) string {
	switch eventType {
	case "goal_home", "goal_away", "kickoff_home", "kickoff_away", "halftime", "fulltime":
		return zoneMiddle
	case "foul_home", "foul_away", "corner", "cross":
		return zoneFinalThird
	case "tackle", "interception":
		if possession == teamHome {
			return zoneOwnThird
		}
		return zoneOwnThird
	case "long_pass":
		if currentZone == zoneOwnBox {
			return zoneMiddle
		}
		if currentZone == zoneOwnThird {
			return zoneFinalThird
		}
		return zoneOppBox
	case "dribble":
		if currentZone == zoneMiddle {
			return zoneFinalThird
		}
		if currentZone == zoneFinalThird {
			return zoneOppBox
		}
	}
	return currentZone
}

func buildFieldSnapshot(cfg DebugMatchConfig, tick int, possession, ballZone, homeMode, awayMode string, homePlayers, awayPlayers []TacticalPlayer) FieldSnapshot {
	rng := rand.New(rand.NewSource(seedForTick(cfg.Seed+404, cfg.MatchID, tick)))
	homeActors := buildActors(homePlayers, true, homeMode, possession, ballZone)
	awayActors := buildActors(awayPlayers, false, awayMode, possession, ballZone)
	ball := ballPoint(possession, ballZone, rng)
	referee := FieldPoint{X: clampInt(ball.X-1+rng.Intn(3), 1, debugFieldWidth-2), Y: clampInt(ball.Y+1, 1, debugFieldHeight-2)}

	return FieldSnapshot{
		Width:      debugFieldWidth,
		Height:     debugFieldHeight,
		HomeName:   defaultTeamName(cfg.Home.Name, true),
		AwayName:   defaultTeamName(cfg.Away.Name, false),
		Ball:       ball,
		Referee:    referee,
		Linesmen:   [2]FieldPoint{{X: clampInt(ball.X-3, 1, debugFieldWidth-2), Y: 1}, {X: clampInt(ball.X+3, 1, debugFieldWidth-2), Y: debugFieldHeight - 2}},
		Home:       homeActors,
		Away:       awayActors,
		Possession: possession,
		BallZone:   ballZone,
		HomeMode:   homeMode,
		AwayMode:   awayMode,
	}
}

func buildActors(players []TacticalPlayer, isHome bool, mode, possession, ballZone string) []DebugActor {
	actors := make([]DebugActor, 0, len(players))
	counts := map[string]int{}

	for _, p := range players {
		canonical := canonicalPosition(p.Position, p.Role)
		counts[canonical]++
		point := playerPoint(canonical, normalizeRole(p.Role), counts[canonical]-1, isHome, mode, possession, ballZone, p.Name)
		actors = append(actors, DebugActor{
			Name:     p.Name,
			Role:     normalizeRole(p.Role),
			Position: canonical,
			Point:    point,
		})
	}

	return actors
}

func playerPoint(position, role string, duplicateIndex int, isHome bool, mode, possession, ballZone, name string) FieldPoint {
	anchor := positionAnchor(position)
	direction := 1
	if !isHome {
		direction = -1
		anchor.X = 100 - anchor.X
	}

	phaseShift := 0
	switch mode {
	case modeOffense:
		phaseShift = 8
	case modeCounterAttack:
		phaseShift = 4
	default:
		phaseShift = -5
	}
	phaseShift = int(float64(phaseShift) * lineShiftFactor(role))

	ballShift := 0
	if (possession == teamHome) == isHome {
		switch ballZone {
		case zoneOwnBox:
			ballShift = -2
		case zoneOwnThird:
			ballShift = 0
		case zoneMiddle:
			ballShift = 2
		case zoneFinalThird:
			ballShift = 5
		case zoneOppBox:
			ballShift = 8
		}
	} else {
		switch ballZone {
		case zoneOwnBox:
			ballShift = 4
		case zoneOwnThird:
			ballShift = 2
		case zoneMiddle:
			ballShift = 0
		case zoneFinalThird:
			ballShift = -3
		case zoneOppBox:
			ballShift = -6
		}
	}

	offsetSeed := stableHash(name + ":" + position)
	yOffset := duplicateYOffset(position, duplicateIndex)
	yOffset += int(offsetSeed%3) - 1

	x := percentToGrid(anchor.X + direction*(phaseShift+ballShift))
	y := percentToLane(anchor.Y + yOffset)

	return FieldPoint{
		X: clampInt(x, 1, debugFieldWidth-2),
		Y: clampInt(y, 1, debugFieldHeight-2),
	}
}

type positionPercent struct {
	X int
	Y int
}

func positionAnchor(position string) positionPercent {
	switch position {
	case "goalkeeper":
		return positionPercent{X: 6, Y: 50}
	case "left_back":
		return positionPercent{X: 21, Y: 18}
	case "right_back":
		return positionPercent{X: 21, Y: 82}
	case "center_back":
		return positionPercent{X: 17, Y: 50}
	case "defensive_midfielder":
		return positionPercent{X: 34, Y: 50}
	case "left_midfielder", "left_wingback", "left_winger":
		return positionPercent{X: 40, Y: 20}
	case "right_midfielder", "right_wingback", "right_winger":
		return positionPercent{X: 40, Y: 80}
	case "central_midfielder":
		return positionPercent{X: 45, Y: 50}
	case "attacking_midfielder":
		return positionPercent{X: 58, Y: 50}
	case "second_striker":
		return positionPercent{X: 66, Y: 50}
	case "striker", "center_forward":
		return positionPercent{X: 76, Y: 50}
	default:
		return positionPercent{X: 44, Y: 50}
	}
}

func canonicalPosition(position, role string) string {
	normalized := normalizeText(position)
	switch normalized {
	case "", "auto":
		return fallbackPositionForRole(role)
	case "goalkeeper", "goleiro", "keeper":
		return "goalkeeper"
	case "zagueiro", "centerback", "centreback", "center_back", "centre_back":
		return "center_back"
	case "lateralesquerdo", "leftback", "left_back":
		return "left_back"
	case "lateraldireito", "rightback", "right_back":
		return "right_back"
	case "volante", "defensivemidfielder", "defensive_midfielder":
		return "defensive_midfielder"
	case "meia", "meiocampista", "centralmidfielder", "central_midfielder":
		return "central_midfielder"
	case "meiaofensivo", "attackingmidfielder", "attacking_midfielder":
		return "attacking_midfielder"
	case "alaesquerda", "pontaesquerda", "leftwinger", "left_winger", "leftmidfielder", "left_midfielder":
		return "left_winger"
	case "aladireita", "pontadireita", "rightwinger", "right_winger", "rightmidfielder", "right_midfielder":
		return "right_winger"
	case "segundoatacante", "secondstriker", "second_striker":
		return "second_striker"
	case "centroavante", "striker", "centerforward", "centreforward", "center_forward", "centre_forward":
		return "center_forward"
	default:
		return fallbackPositionForRole(role)
	}
}

func fallbackPositionForRole(role string) string {
	switch normalizeRole(role) {
	case "goalkeeper":
		return "goalkeeper"
	case "defender":
		return "center_back"
	case "fullback":
		return "left_back"
	case "forward":
		return "center_forward"
	default:
		return "central_midfielder"
	}
}

func normalizeText(value string) string {
	clean := strings.NewReplacer(
		" ", "",
		"-", "",
		"_", "",
		"/", "",
		"á", "a",
		"à", "a",
		"â", "a",
		"ã", "a",
		"é", "e",
		"ê", "e",
		"í", "i",
		"ó", "o",
		"ô", "o",
		"õ", "o",
		"ú", "u",
		"ç", "c",
	).Replace(strings.ToLower(strings.TrimSpace(value)))
	return clean
}

func lineShiftFactor(role string) float64 {
	switch role {
	case "goalkeeper":
		return 0.2
	case "defender":
		return 0.55
	case "fullback":
		return 0.85
	case "forward":
		return 1.15
	default:
		return 1.0
	}
}

func duplicateYOffset(position string, duplicateIndex int) int {
	patterns := map[string][]int{
		"center_back":         {-10, 10, 0},
		"central_midfielder":  {-14, 0, 14},
		"center_forward":      {-8, 8, 0},
		"second_striker":      {-6, 6, 0},
		"left_winger":         {-5, 5},
		"right_winger":        {-5, 5},
		"defensive_midfielder": {-8, 8},
	}
	options, ok := patterns[position]
	if !ok || len(options) == 0 {
		return duplicateIndex * 2
	}
	return options[duplicateIndex%len(options)]
}

func ballPoint(possession, zone string, rng *rand.Rand) FieldPoint {
	xPercent := 50
	switch zone {
	case zoneOwnBox:
		xPercent = 12
	case zoneOwnThird:
		xPercent = 24
	case zoneMiddle:
		xPercent = 50
	case zoneFinalThird:
		xPercent = 72
	case zoneOppBox:
		xPercent = 84
	}
	if possession == teamAway {
		xPercent = 100 - xPercent
	}

	return FieldPoint{
		X: clampInt(percentToGrid(xPercent), 1, debugFieldWidth-2),
		Y: clampInt(percentToLane(25+rng.Intn(51)), 1, debugFieldHeight-2),
	}
}

func percentToGrid(percent int) int {
	return 1 + (percent*(debugFieldWidth-3))/100
}

func percentToLane(percent int) int {
	return 1 + (percent*(debugFieldHeight-3))/100
}

func baseFieldCell(x, y int) string {
	if (x == 0 || x == debugFieldWidth-1) && (y == 0 || y == debugFieldHeight-1) {
		return "+"
	}
	if y == 0 || y == debugFieldHeight-1 {
		return "-"
	}
	if x == 0 || x == debugFieldWidth-1 {
		return "|"
	}
	if x == debugFieldWidth/2 && y == debugFieldHeight/2 {
		return "+"
	}
	if x == debugFieldWidth/2 {
		return ":"
	}
	if x == 6 || x == debugFieldWidth-7 {
		if y >= 5 && y <= debugFieldHeight-6 {
			return ":"
		}
	}
	return "."
}

func snapshotTeamName(snapshot DebugSnapshot, home bool) string {
	if home {
		if snapshot.Field.HomeName != "" {
			return snapshot.Field.HomeName
		}
		return "Casa"
	}
	if snapshot.Field.AwayName != "" {
		return snapshot.Field.AwayName
	}
	return "Fora"
}

func nextPossession(current, eventType string) string {
	switch eventType {
	case "tackle", "interception", "goal_home", "goal_away", "foul_home", "foul_away":
		if current == teamHome {
			return teamAway
		}
		return teamHome
	default:
		return current
	}
}

func defaultTeamName(name string, home bool) string {
	trimmed := strings.TrimSpace(name)
	if trimmed != "" {
		return trimmed
	}
	if home {
		return "Casa"
	}
	return "Fora"
}

func stableHash(value string) uint64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(value))
	return hash.Sum64()
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}