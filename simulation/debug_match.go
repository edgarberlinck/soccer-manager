package simulation

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"sort"
	"strings"

	"manager/game/internal/domain/player"

	"github.com/google/uuid"
)

const (
	debugFieldWidth       = 41
	debugFieldHeight      = 17
	maxDebugSubstitutions = 5

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
	ID       string
	Team     string
	Name     string
	Role     string
	Position string
	Point    FieldPoint
	Stamina  int
	Starter  bool
	Active   bool
}

type FieldCellOccupant struct {
	Kind    string
	Team    string
	ActorID string
	Label   string
}

type FieldCell struct {
	X         int
	Y         int
	Base      string
	Occupants []FieldCellOccupant
}

type DebugSubstitution struct {
	Tick          int
	Team          string
	OutPlayerID   string
	OutPlayerName string
	InPlayerID    string
	InPlayerName  string
	Reason        string
}

type SquadMemberSnapshot struct {
	ID           string
	Team         string
	Name         string
	Role         string
	Position     string
	Stamina      int
	Starter      bool
	Active       bool
	SubbedInTick int
	SubbedOutTick int
	Point        *FieldPoint
}

type FieldSnapshot struct {
	Width             int
	Height            int
	HomeName          string
	AwayName          string
	Ball              FieldPoint
	Referee           FieldPoint
	Linesmen          [2]FieldPoint
	Home              []DebugActor
	Away              []DebugActor
	HomeSquad         []SquadMemberSnapshot
	AwaySquad         []SquadMemberSnapshot
	Substitutions     []DebugSubstitution
	HomeSubsUsed      int
	AwaySubsUsed      int
	Possession        string
	BallZone          string
	HomeMode          string
	AwayMode          string
	Matrix            [][]FieldCell
	LowestHomeStamina int
	LowestAwayStamina int
}

type DebugSnapshot struct {
	Outcome TickOutcome
	Field   FieldSnapshot
}

type debugPlayerState struct {
	ID            string
	Team          string
	Player        TacticalPlayer
	BaseAttributes player.Attributes
	Stamina       float64
	Starter       bool
	Active        bool
	SubbedInTick  int
	SubbedOutTick int
	LastPoint     FieldPoint
}

func SimulateDebugMatch(cfg DebugMatchConfig) []DebugSnapshot {
	homeStates := newDebugPlayerStates(cfg.Home.Players, cfg.Home.ClubID, teamHome)
	awayStates := newDebugPlayerStates(cfg.Away.Players, cfg.Away.ClubID, teamAway)

	possession := ""
	ballZone := zoneMiddle
	homeScore := 0
	awayScore := 0
	homeSubsUsed := 0
	awaySubsUsed := 0

	snapshots := make([]DebugSnapshot, 0, regulationTicks)

	for tick := 1; tick <= regulationTicks; tick++ {
		homeSubs, homeUsed := applyAutoSubstitutions(tick, homeStates, homeSubsUsed)
		awaySubs, awayUsed := applyAutoSubstitutions(tick, awayStates, awaySubsUsed)
		homeSubsUsed = homeUsed
		awaySubsUsed = awayUsed

		outcome := PlayMatchTick(PlayMatchTickInput{
			MatchID:          cfg.MatchID,
			CurrentTick:      tick,
			Seed:             cfg.Seed,
			HomeClubID:       cfg.Home.ClubID,
			AwayClubID:       cfg.Away.ClubID,
			HomeScore:        homeScore,
			AwayScore:        awayScore,
			PossessionTeam:   possession,
			BallZone:         ballZone,
			EnvironmentNoise: true,
			HomePlayers:      activeLineup(homeStates),
			AwayPlayers:      activeLineup(awayStates),
		})

		homeScore = outcome.HomeScore
		awayScore = outcome.AwayScore

		statePossession, stateZone := resolveDebugState(outcome, possession, ballZone)
		statePossession = nextPossession(statePossession, outcome.EventType)
		stateZone = advanceBallZone(stateZone, statePossession, outcome.EventType)
		homeMode, awayMode := resolveModes(statePossession, stateZone)

		applyFatigue(homeStates, homeMode, statePossession == teamHome, tick)
		applyFatigue(awayStates, awayMode, statePossession == teamAway, tick)

		field := buildFieldSnapshot(cfg, tick, statePossession, stateZone, homeMode, awayMode, homeStates, awayStates, homeSubsUsed, awaySubsUsed, append(homeSubs, awaySubs...))

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

		snapshots = append(snapshots, DebugSnapshot{Outcome: outcome, Field: field})

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
			grid[y][x] = snapshot.Field.Matrix[y][x].Base
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
	lines = append(lines, fmt.Sprintf("Stamina minima: casa=%d fora=%d | substituicoes: %d/%d e %d/%d", snapshot.Field.LowestHomeStamina, snapshot.Field.LowestAwayStamina, snapshot.Field.HomeSubsUsed, maxDebugSubstitutions, snapshot.Field.AwaySubsUsed, maxDebugSubstitutions))
	if len(snapshot.Field.Substitutions) > 0 {
		parts := make([]string, 0, len(snapshot.Field.Substitutions))
		for _, sub := range snapshot.Field.Substitutions {
			parts = append(parts, fmt.Sprintf("%s: saiu %s, entrou %s (%s)", sub.Team, sub.OutPlayerName, sub.InPlayerName, sub.Reason))
		}
		lines = append(lines, "Substituicoes: "+strings.Join(parts, " | "))
	}
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

func buildFieldSnapshot(cfg DebugMatchConfig, tick int, possession, ballZone, homeMode, awayMode string, homeStates, awayStates []debugPlayerState, homeSubsUsed, awaySubsUsed int, substitutions []DebugSubstitution) FieldSnapshot {
	rng := rand.New(rand.NewSource(seedForTick(cfg.Seed+404, cfg.MatchID, tick)))
	homeActors := buildActors(homeStates, true, homeMode, possession, ballZone)
	awayActors := buildActors(awayStates, false, awayMode, possession, ballZone)
	ball := ballPoint(possession, ballZone, rng)
	referee := FieldPoint{X: clampInt(ball.X-1+rng.Intn(3), 1, debugFieldWidth-2), Y: clampInt(ball.Y+1, 1, debugFieldHeight-2)}

	snapshot := FieldSnapshot{
		Width:      debugFieldWidth,
		Height:     debugFieldHeight,
		HomeName:   defaultTeamName(cfg.Home.Name, true),
		AwayName:   defaultTeamName(cfg.Away.Name, false),
		Ball:       ball,
		Referee:    referee,
		Linesmen:   [2]FieldPoint{{X: clampInt(ball.X-3, 1, debugFieldWidth-2), Y: 1}, {X: clampInt(ball.X+3, 1, debugFieldWidth-2), Y: debugFieldHeight - 2}},
		Home:       homeActors,
		Away:       awayActors,
		HomeSquad:  buildSquadSnapshot(homeStates),
		AwaySquad:  buildSquadSnapshot(awayStates),
		Substitutions: substitutions,
		HomeSubsUsed: homeSubsUsed,
		AwaySubsUsed: awaySubsUsed,
		Possession: possession,
		BallZone:   ballZone,
		HomeMode:   homeMode,
		AwayMode:   awayMode,
		LowestHomeStamina: lowestStamina(homeStates),
		LowestAwayStamina: lowestStamina(awayStates),
	}
	snapshot.Matrix = buildFieldMatrix(snapshot)
	return snapshot
}

func buildActors(states []debugPlayerState, isHome bool, mode, possession, ballZone string) []DebugActor {
	actors := make([]DebugActor, 0, 11)
	counts := map[string]int{}

	for i := range states {
		if !states[i].Active {
			continue
		}
		p := states[i].currentTacticalPlayer()
		canonical := canonicalPosition(p.Position, p.Role)
		counts[canonical]++
		point := playerPoint(canonical, normalizeRole(p.Role), counts[canonical]-1, isHome, mode, possession, ballZone, p.Name)
		states[i].LastPoint = point
		actors = append(actors, DebugActor{ID: states[i].ID, Team: states[i].Team, Name: p.Name, Role: normalizeRole(p.Role), Position: canonical, Point: point, Stamina: clampStatValue(int(math.Round(states[i].Stamina))), Starter: states[i].Starter, Active: true})
	}

	return actors
}

func buildSquadSnapshot(states []debugPlayerState) []SquadMemberSnapshot {
	snapshot := make([]SquadMemberSnapshot, 0, len(states))
	for i := range states {
		var point *FieldPoint
		if states[i].Active {
			p := states[i].LastPoint
			point = &p
		}
		snapshot = append(snapshot, SquadMemberSnapshot{
			ID: states[i].ID,
			Team: states[i].Team,
			Name: states[i].Player.Name,
			Role: normalizeRole(states[i].Player.Role),
			Position: canonicalPosition(states[i].Player.Position, states[i].Player.Role),
			Stamina: clampStatValue(int(math.Round(states[i].Stamina))),
			Starter: states[i].Starter,
			Active: states[i].Active,
			SubbedInTick: states[i].SubbedInTick,
			SubbedOutTick: states[i].SubbedOutTick,
			Point: point,
		})
	}
	return snapshot
}

func buildFieldMatrix(snapshot FieldSnapshot) [][]FieldCell {
	matrix := make([][]FieldCell, debugFieldHeight)
	for y := 0; y < debugFieldHeight; y++ {
		matrix[y] = make([]FieldCell, debugFieldWidth)
		for x := 0; x < debugFieldWidth; x++ {
			matrix[y][x] = FieldCell{X: x, Y: y, Base: baseFieldCell(x, y)}
		}
	}

	place := func(point FieldPoint, occupant FieldCellOccupant) {
		if point.Y < 0 || point.Y >= debugFieldHeight || point.X < 0 || point.X >= debugFieldWidth {
			return
		}
		matrix[point.Y][point.X].Occupants = append(matrix[point.Y][point.X].Occupants, occupant)
	}

	for _, actor := range snapshot.Home {
		place(actor.Point, FieldCellOccupant{Kind: "player", Team: actor.Team, ActorID: actor.ID, Label: actor.Name})
	}
	for _, actor := range snapshot.Away {
		place(actor.Point, FieldCellOccupant{Kind: "player", Team: actor.Team, ActorID: actor.ID, Label: actor.Name})
	}
	place(snapshot.Ball, FieldCellOccupant{Kind: "ball", Label: "ball"})
	place(snapshot.Referee, FieldCellOccupant{Kind: "referee", Label: "juiz"})
	place(snapshot.Linesmen[0], FieldCellOccupant{Kind: "linesman", Label: "bandeirinha_1"})
	place(snapshot.Linesmen[1], FieldCellOccupant{Kind: "linesman", Label: "bandeirinha_2"})

	return matrix
}

func newDebugPlayerStates(players []TacticalPlayer, teamID uuid.UUID, team string) []debugPlayerState {
	ensured := ensureTeamPlayers(players, teamID)
	states := make([]debugPlayerState, 0, len(ensured))
	for index, p := range ensured {
		stamina := float64(p.Attributes.FisicalStatus)
		if stamina <= 0 {
			stamina = 78
		}
		states = append(states, debugPlayerState{
			ID: stablePlayerID(team, p.Name, index),
			Team: team,
			Player: p,
			BaseAttributes: p.Attributes,
			Stamina: stamina,
			Starter: index < 11,
			Active: index < 11,
		})
	}
	return states
}

func activeLineup(states []debugPlayerState) []TacticalPlayer {
	lineup := make([]TacticalPlayer, 0, 11)
	for i := range states {
		if states[i].Active {
			lineup = append(lineup, states[i].currentTacticalPlayer())
		}
	}
	return lineup
}

func applyAutoSubstitutions(tick int, states []debugPlayerState, used int) ([]DebugSubstitution, int) {
	if tick < 55 || used >= maxDebugSubstitutions {
		return nil, used
	}

	bestOut := -1
	bestIn := -1
	bestNeed := -1.0

	for outIdx := range states {
		if !states[outIdx].Active {
			continue
		}
		benchIdx, need := bestBenchReplacement(states, outIdx)
		if benchIdx < 0 || need <= bestNeed {
			continue
		}
		bestOut = outIdx
		bestIn = benchIdx
		bestNeed = need
	}

	if bestOut < 0 || bestIn < 0 || bestNeed < 0.16 {
		return nil, used
	}

	reason := "baixo impacto"
	if states[bestOut].Stamina < 58 {
		reason = "desgaste"
	}

	states[bestOut].Active = false
	states[bestOut].SubbedOutTick = tick
	states[bestIn].Active = true
	states[bestIn].SubbedInTick = tick
	if states[bestIn].Stamina < 65 {
		states[bestIn].Stamina = 65
	}

	sub := DebugSubstitution{
		Tick: tick,
		Team: states[bestOut].Team,
		OutPlayerID: states[bestOut].ID,
		OutPlayerName: states[bestOut].Player.Name,
		InPlayerID: states[bestIn].ID,
		InPlayerName: states[bestIn].Player.Name,
		Reason: reason,
	}

	return []DebugSubstitution{sub}, used + 1
}

func bestBenchReplacement(states []debugPlayerState, outIdx int) (int, float64) {
	out := states[outIdx]
	outContribution := playerContribution(out.currentTacticalPlayer()) * (out.Stamina / 100)
	bestIdx := -1
	bestGain := -1.0
	for inIdx := range states {
		if states[inIdx].Active {
			continue
		}
		fit := substitutionFit(out, states[inIdx])
		if fit <= 0 {
			continue
		}
		benchContribution := playerContribution(states[inIdx].currentTacticalPlayer()) * fit
		gain := benchContribution - outContribution
		if out.Stamina < 58 {
			gain += (58 - out.Stamina) / 100
		}
		if gain > bestGain {
			bestGain = gain
			bestIdx = inIdx
		}
	}
	return bestIdx, bestGain
}

func substitutionFit(out, in debugPlayerState) float64 {
	if normalizeRole(out.Player.Role) == normalizeRole(in.Player.Role) {
		return 1.0
	}
	if canonicalPosition(out.Player.Position, out.Player.Role) == canonicalPosition(in.Player.Position, in.Player.Role) {
		return 0.92
	}
	if lineGroup(out.Player.Role) == lineGroup(in.Player.Role) {
		return 0.7
	}
	return 0
}

func lineGroup(role string) string {
	switch normalizeRole(role) {
	case "goalkeeper":
		return "gk"
	case "defender", "fullback":
		return "defense"
	case "forward":
		return "attack"
	default:
		return "midfield"
	}
}

func playerContribution(p TacticalPlayer) float64 {
	a := p.Attributes
	return float64(a.Pace+a.Passing+a.Shooting+a.Explosao+a.Fisico+a.Habilidade+a.Finalizacao+a.Dominio) / 800
}

func applyFatigue(states []debugPlayerState, mode string, hasPossession bool, tick int) {
	for i := range states {
		if !states[i].Active {
			continue
		}
		drain := 0.45 + roleDrain(states[i].Player.Role)
		if hasPossession {
			drain += 0.18
		}
		switch mode {
		case modeOffense:
			drain += 0.22
		case modeCounterAttack:
			drain += 0.18
		default:
			drain += 0.12
		}
		if tick > 75 {
			drain += 0.1
		}
		states[i].Stamina = clampFloat(states[i].Stamina-drain, 18, 99)
	}
}

func roleDrain(role string) float64 {
	switch normalizeRole(role) {
	case "goalkeeper":
		return 0.05
	case "defender":
		return 0.14
	case "fullback":
		return 0.2
	case "forward":
		return 0.22
	default:
		return 0.18
	}
}

func lowestStamina(states []debugPlayerState) int {
	lowest := 99
	for i := range states {
		if !states[i].Active {
			continue
		}
		value := clampStatValue(int(math.Round(states[i].Stamina)))
		if value < lowest {
			lowest = value
		}
	}
	if lowest == 99 && len(states) == 0 {
		return 0
	}
	return lowest
}

func (state debugPlayerState) currentTacticalPlayer() TacticalPlayer {
	playerCopy := state.Player
	playerCopy.Attributes = adjustedAttributes(state.BaseAttributes, state.Stamina)
	return playerCopy
}

func adjustedAttributes(base player.Attributes, stamina float64) player.Attributes {
	fatigueFactor := clampFloat(stamina/100, 0.58, 1)
	softFactor := 0.78 + 0.22*fatigueFactor
	base.FisicalStatus = clampStatValue(int(math.Round(stamina)))
	base.Pace = scaleStat(base.Pace, fatigueFactor)
	base.Explosao = scaleStat(base.Explosao, fatigueFactor)
	base.Fisico = scaleStat(base.Fisico, 0.82+0.18*fatigueFactor)
	base.Passing = scaleStat(base.Passing, softFactor)
	base.Dominio = scaleStat(base.Dominio, softFactor)
	base.Finalizacao = scaleStat(base.Finalizacao, softFactor)
	base.Habilidade = scaleStat(base.Habilidade, softFactor)
	return base
}

func scaleStat(value int, factor float64) int {
	return clampStatValue(int(math.Round(float64(value) * factor)))
}

func clampStatValue(value int) int {
	if value < 1 {
		return 1
	}
	if value > 99 {
		return 99
	}
	return value
}

func stablePlayerID(team, name string, index int) string {
	return fmt.Sprintf("%s-%x-%d", team, stableHash(name), index)
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

func sortSquadSnapshots(members []SquadMemberSnapshot) {
	sort.SliceStable(members, func(i, j int) bool {
		if members[i].Active != members[j].Active {
			return members[i].Active
		}
		if members[i].Starter != members[j].Starter {
			return members[i].Starter
		}
		return members[i].Name < members[j].Name
	})
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