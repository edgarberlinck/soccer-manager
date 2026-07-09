package simulation

import (
	"fmt"
	"hash/fnv"
	"manager/game/internal/domain/club"
	"manager/game/internal/domain/match"
	"manager/game/internal/domain/player"
	"math"
	"math/rand"
	"strings"

	"github.com/google/uuid"
)

const regulationTicks = 90

const (
	teamHome = "home"
	teamAway = "away"
)

const (
	modeOffense       = "offense"
	modeDefense       = "defense"
	modeCounterAttack = "counter_attack"
)

const (
	zoneOwnBox    = "own_box"
	zoneOwnThird  = "own_third"
	zoneMiddle    = "middle"
	zoneFinalThird = "final_third"
	zoneOppBox    = "opp_box"
)

// TacticalPlayer carrega o recorte minimo de um jogador necessario para a
// simulacao tática por tick.
//
// A Role define o comportamento de movimentacao em cada fase:
// - goalkeeper
// - defender
// - fullback
// - midfielder
// - forward
//
// Se Role vier vazio, a simulacao assume "midfielder".
type TacticalPlayer struct {
	Name       string
	Role       string
	Position   string
	Attributes player.Attributes
}

// PlayMatchTickInput define o estado minimo para processar um tick.
//
// Campos obrigatorios (compativeis com a versao anterior):
// - MatchID, CurrentTick, Seed, HomeClubID, AwayClubID, HomeScore, AwayScore.
//
// Campos opcionais para controle fino da simulacao tática:
// - PossessionTeam: "home", "away" ou vazio para decisao automatica.
// - BallZone: "own_box", "own_third", "middle", "final_third", "opp_box".
//             A zona e sempre relativa ao time que esta com a posse.
// - HomePlayers/AwayPlayers: elenco do tick. Se vazio, a simulacao usa
//   fatores agregados por ID de clube para manter o comportamento anterior.
type PlayMatchTickInput struct {
	MatchID         uuid.UUID
	CurrentTick     int
	Seed            int64
	HomeClubID      uuid.UUID
	AwayClubID      uuid.UUID
	HomeScore       int
	AwayScore       int
	PossessionTeam  string
	BallZone        string
	EnvironmentNoise bool
	HomePlayers     []TacticalPlayer
	AwayPlayers     []TacticalPlayer
}

// TickOutcome descreve o resultado completo de um tick.
//
// Alem de evento/placar, expomos telemetria tática para facilitar debug,
// ajuste de balanceamento e testes de regressao da simulacao.
type TickOutcome struct {
	Tick          int
	NextTick      int
	EventType     string
	Description   string
	HomeScore     int
	AwayScore     int
	IsFinished    bool
	PossessionTeam string
	BallZone      string
	HomeMode      string
	AwayMode      string
	HomeAdvance   int
	AwayAdvance   int
	HomePressure  int
	AwayPressure  int
}

type tacticalFrame struct {
	possessionTeam string
	ballZone       string
	homeMode       string
	awayMode       string
	homeMetrics    teamMetrics
	awayMetrics    teamMetrics
	homeGoalChance float64
	awayGoalChance float64
}

type teamMetrics struct {
	Advance      int
	Pressure     int
	AttackPower  float64
	DefensePower float64
	InjuryRisk   float64
	FoulRisk     float64
}

// PlayMatchTick executa um minuto (tick) de partida com comportamento
// deterministico para a mesma combinacao de entrada.
//
// Regras gerais implementadas:
// 1. O time com posse entra em modo ofensivo, exceto se a bola estiver na
//    propria area/terco defensivo, onde a posse e tratada como contra-ataque.
// 2. O time sem posse fica em modo defensivo, com foco em pressao/marcacao.
// 3. A movimentacao depende dos atributos e da funcao do jogador.
// 4. Baixo fisical_status eleva risco de lesao.
// 5. Baixo temperamento eleva risco de falta.
func PlayMatchTick(input PlayMatchTickInput) TickOutcome {
	tick := normalizeTick(input.CurrentTick)
	rng := rand.New(rand.NewSource(seedForTick(input.Seed, input.MatchID, tick)))

	outcome := TickOutcome{
		Tick:       tick,
		NextTick:   nextTick(tick),
		HomeScore:  input.HomeScore,
		AwayScore:  input.AwayScore,
		IsFinished: tick >= regulationTicks,
	}

	if tick == 1 {
		if resolvePossessionTeam(input.PossessionTeam, rng) == teamHome {
			outcome.EventType = "kickoff_home"
			outcome.Description = "Inicio de jogo: time da casa sai com a bola."
		} else {
			outcome.EventType = "kickoff_away"
			outcome.Description = "Inicio de jogo: time visitante sai com a bola."
		}
		return outcome
	}

	if tick == 45 {
		outcome.EventType = "halftime"
		outcome.Description = fmt.Sprintf("Fim do primeiro tempo: %d x %d.", outcome.HomeScore, outcome.AwayScore)
		return outcome
	}

	frame := buildTacticalFrame(input, rng, tick)
	outcome.PossessionTeam = frame.possessionTeam
	outcome.BallZone = frame.ballZone
	outcome.HomeMode = frame.homeMode
	outcome.AwayMode = frame.awayMode
	outcome.HomeAdvance = frame.homeMetrics.Advance
	outcome.AwayAdvance = frame.awayMetrics.Advance
	outcome.HomePressure = frame.homeMetrics.Pressure
	outcome.AwayPressure = frame.awayMetrics.Pressure

	if frame.homeMetrics.InjuryRisk > 0 && rng.Float64() < frame.homeMetrics.InjuryRisk {
		outcome.EventType = "injury_home"
		outcome.Description = fmt.Sprintf("Tick %d: jogador do mandante sente lesao apos alta carga fisica.", tick)
		return finalizeAtFulltime(outcome)
	}
	if frame.awayMetrics.InjuryRisk > 0 && rng.Float64() < frame.awayMetrics.InjuryRisk {
		outcome.EventType = "injury_away"
		outcome.Description = fmt.Sprintf("Tick %d: jogador visitante sente lesao apos alta carga fisica.", tick)
		return finalizeAtFulltime(outcome)
	}

	draw := rng.Float64()

	if draw < frame.homeGoalChance {
		outcome.HomeScore++
		outcome.EventType = "goal_home"
		outcome.Description = fmt.Sprintf("Gol do mandante no tick %d. Placar: %d x %d.", tick, outcome.HomeScore, outcome.AwayScore)
		return finalizeAtFulltime(outcome)
	}
	if draw < frame.homeGoalChance+frame.awayGoalChance {
		outcome.AwayScore++
		outcome.EventType = "goal_away"
		outcome.Description = fmt.Sprintf("Gol do visitante no tick %d. Placar: %d x %d.", tick, outcome.HomeScore, outcome.AwayScore)
		return finalizeAtFulltime(outcome)
	}

	if frame.homeMetrics.FoulRisk > 0 && rng.Float64() < frame.homeMetrics.FoulRisk {
		outcome.EventType = "foul_home"
		outcome.Description = fmt.Sprintf("Tick %d: falta do mandante por excesso de contato na marcacao.", tick)
		return finalizeAtFulltime(outcome)
	}
	if frame.awayMetrics.FoulRisk > 0 && rng.Float64() < frame.awayMetrics.FoulRisk {
		outcome.EventType = "foul_away"
		outcome.Description = fmt.Sprintf("Tick %d: falta do visitante por excesso de contato na marcacao.", tick)
		return finalizeAtFulltime(outcome)
	}

	outcome.EventType, outcome.Description = nonGoalEventWithContext(rng, tick, frame)
	return finalizeAtFulltime(outcome)
}

func finalizeAtFulltime(outcome TickOutcome) TickOutcome {
	if outcome.Tick == regulationTicks {
		outcome.EventType = "fulltime"
		outcome.Description = fmt.Sprintf("Fim de jogo: %d x %d.", outcome.HomeScore, outcome.AwayScore)
		outcome.IsFinished = true
	}
	return outcome
}

func buildTacticalFrame(input PlayMatchTickInput, rng *rand.Rand, tick int) tacticalFrame {
	possession := resolvePossessionTeam(input.PossessionTeam, rng)

	homePlayers := ensureTeamPlayers(input.HomePlayers, input.HomeClubID)
	awayPlayers := ensureTeamPlayers(input.AwayPlayers, input.AwayClubID)

	homeBase := evaluateTeamMetrics(homePlayers, modeDefense)
	awayBase := evaluateTeamMetrics(awayPlayers, modeDefense)

	zone := resolveBallZone(input.BallZone, possession, rng, homeBase.Advance, awayBase.Advance, input.EnvironmentNoise)
	homeMode, awayMode := resolveModes(possession, zone)

	homeMetrics := evaluateTeamMetrics(homePlayers, homeMode)
	awayMetrics := evaluateTeamMetrics(awayPlayers, awayMode)

	homeGoalChance, awayGoalChance := contextualGoalChances(input.HomeClubID, input.AwayClubID, possession, zone, homeMetrics, awayMetrics)

	if tick >= regulationTicks {
		homeGoalChance = 0
		awayGoalChance = 0
	}

	return tacticalFrame{
		possessionTeam: possession,
		ballZone:       zone,
		homeMode:       homeMode,
		awayMode:       awayMode,
		homeMetrics:    homeMetrics,
		awayMetrics:    awayMetrics,
		homeGoalChance: homeGoalChance,
		awayGoalChance: awayGoalChance,
	}
}

func resolvePossessionTeam(requested string, rng *rand.Rand) string {
	switch strings.ToLower(strings.TrimSpace(requested)) {
	case teamHome:
		return teamHome
	case teamAway:
		return teamAway
	default:
		if rng.Intn(2) == 0 {
			return teamHome
		}
		return teamAway
	}
}

func resolveBallZone(requested, possession string, rng *rand.Rand, homeAdvance, awayAdvance int, environmentNoise bool) string {
	if isValidZone(requested) {
		if environmentNoise {
			return applyBallDrift(requested, possession, rng, homeAdvance, awayAdvance)
		}
		return requested
	}

	pressureDelta := float64(homeAdvance-awayAdvance) / 120
	if possession == teamAway {
		pressureDelta *= -1
	}

	roll := rng.Float64() + clampFloat(pressureDelta, -0.18, 0.18)
	switch {
	case roll < 0.12:
		return zoneOwnBox
	case roll < 0.32:
		return zoneOwnThird
	case roll < 0.68:
		return zoneMiddle
	case roll < 0.9:
		return zoneFinalThird
	default:
		return zoneOppBox
	}
}

func applyBallDrift(zone, possession string, rng *rand.Rand, homeAdvance, awayAdvance int) string {
	zones := []string{zoneOwnBox, zoneOwnThird, zoneMiddle, zoneFinalThird, zoneOppBox}
	index := 0
	for i, candidate := range zones {
		if candidate == zone {
			index = i
			break
		}
	}

	pressureDelta := clampFloat(float64(homeAdvance-awayAdvance)/220, -0.06, 0.06)
	if possession == teamAway {
		pressureDelta *= -1
	}

	backwardChance := clampFloat(0.12-pressureDelta, 0.05, 0.18)
	forwardChance := clampFloat(0.12+pressureDelta, 0.05, 0.18)
	drift := rng.Float64()

	if drift < backwardChance {
		index--
	} else if drift > 1-forwardChance {
		index++
	}

	if index < 0 {
		index = 0
	}
	if index >= len(zones) {
		index = len(zones) - 1
	}
	return zones[index]
}

func resolveModes(possession, zone string) (string, string) {
	if possession == teamHome {
		if zone == zoneOwnBox || zone == zoneOwnThird {
			return modeCounterAttack, modeDefense
		}
		return modeOffense, modeDefense
	}
	if zone == zoneOwnBox || zone == zoneOwnThird {
		return modeDefense, modeCounterAttack
	}
	return modeDefense, modeOffense
}

func ensureTeamPlayers(players []TacticalPlayer, teamID uuid.UUID) []TacticalPlayer {
	if len(players) > 0 {
		return players
	}

	seed := float64(teamSkillFactor(teamID) * 1000)
	base := int(clampFloat(72+seed, 55, 89))

	mk := func(role string, delta int) TacticalPlayer {
		v := int16(clampFloat(float64(base+delta), 40, 99))
		return TacticalPlayer{
			Role: role,
			Attributes: player.Attributes{
				Pace:          int(v),
				Passing:       int(v),
				Shooting:      int(v),
				Altura:        178,
				Peso:          75,
				Impulso:       int(v),
				Explosao:      int(v),
				Fisico:        int(v),
				FisicalStatus: int(v),
				Cabeceio:      int(v),
				Cruzamento:    int(v),
				Habilidade:    int(v),
				Finalizacao:   int(v),
				Dominio:       int(v),
				Temperamento:  int(v),
			},
		}
	}

	return []TacticalPlayer{
		mk("goalkeeper", -4),
		mk("defender", -1),
		mk("defender", 0),
		mk("fullback", 1),
		mk("fullback", 1),
		mk("midfielder", 2),
		mk("midfielder", 2),
		mk("midfielder", 1),
		mk("forward", 3),
		mk("forward", 2),
		mk("forward", 1),
	}
}

func evaluateTeamMetrics(players []TacticalPlayer, mode string) teamMetrics {
	if len(players) == 0 {
		return teamMetrics{}
	}

	var advance, pressure float64
	var attackPower, defensePower float64
	var injuryRisk, foulRisk float64

	for _, p := range players {
		role := normalizeRole(p.Role)
		a := p.Attributes

		pace := normalizeSkill(a.Pace)
		passing := normalizeSkill(a.Passing)
		shooting := normalizeSkill(a.Shooting)
		jump := normalizeSkill(a.Impulso)
		explosion := normalizeSkill(a.Explosao)
		physical := normalizeSkill(a.Fisico)
		physicalStatus := normalizeSkill(a.FisicalStatus)
		heading := normalizeSkill(a.Cabeceio)
		crossing := normalizeSkill(a.Cruzamento)
		skill := normalizeSkill(a.Habilidade)
		finishing := normalizeSkill(a.Finalizacao)
		control := normalizeSkill(a.Dominio)
		temper := normalizeSkill(a.Temperamento)

		heightScore := clampFloat((float64(a.Altura)-150)/50, 0, 1)
		weightScore := clampFloat((float64(a.Peso)-55)/45, 0, 1)

		athleticBoost := clampFloat((explosion+physical+heightScore+jump+weightScore)/5, 0, 1)
		technicalBoost := clampFloat((passing+skill+crossing+finishing+control+heading+shooting+pace)/8, 0, 1)

		offRole, defRole, pressureRole := roleWeights(role)
		offMode, defMode, pressMode := modeWeights(mode)

		advance += 100 * athleticBoost * technicalBoost * offRole * offMode
		pressure += 100 * (0.55*athleticBoost+0.45*physical) * pressureRole * pressMode

		attackPower += (0.25*finishing + 0.2*skill + 0.2*control + 0.15*passing + 0.1*crossing + 0.1*heading) * offRole * offMode
		defensePower += (0.45*physical + 0.2*pace + 0.2*control + 0.15*heightScore) * defRole * defMode

		exertion := clampFloat((athleticBoost+technicalBoost)/2, 0, 1)
		injuryRisk += clampFloat((1-physicalStatus)*(0.08+0.12*exertion*offMode), 0, 0.2)
		foulRisk += clampFloat((1-temper)*(0.04+0.1*pressMode), 0, 0.16)
	}

	sz := float64(len(players))
	return teamMetrics{
		Advance:      int(math.Round(advance / sz)),
		Pressure:     int(math.Round(pressure / sz)),
		AttackPower:  attackPower / sz,
		DefensePower: defensePower / sz,
		InjuryRisk:   clampFloat(injuryRisk/sz, 0.001, 0.16),
		FoulRisk:     clampFloat(foulRisk/sz, 0.002, 0.13),
	}
}

func contextualGoalChances(homeClubID, awayClubID uuid.UUID, possession, zone string, home, away teamMetrics) (float64, float64) {
	baseHome, baseAway := goalChances(homeClubID, awayClubID)

	zoneBoost := map[string]float64{
		zoneOwnBox:    -0.45,
		zoneOwnThird:  -0.2,
		zoneMiddle:    0,
		zoneFinalThird: 0.18,
		zoneOppBox:    0.32,
	}[zone]

	homeBoost := (home.AttackPower-away.DefensePower)*0.05 + float64(home.Advance-away.Pressure)/2500
	awayBoost := (away.AttackPower-home.DefensePower)*0.05 + float64(away.Advance-home.Pressure)/2500

	if possession == teamHome {
		homeBoost += zoneBoost
		awayBoost -= zoneBoost * 0.45
	} else {
		awayBoost += zoneBoost
		homeBoost -= zoneBoost * 0.45
	}

	homeChance := clampFloat(baseHome*(1+homeBoost), 0.003, 0.16)
	awayChance := clampFloat(baseAway*(1+awayBoost), 0.003, 0.16)

	return homeChance, awayChance
}

func roleWeights(role string) (offense float64, defense float64, pressure float64) {
	switch role {
	case "goalkeeper":
		return 0.2, 1.2, 0.4
	case "defender":
		return 0.55, 1.15, 0.85
	case "fullback":
		return 1.1, 1.0, 1.05
	case "midfielder":
		return 1.0, 1.0, 1.1
	case "forward":
		return 1.25, 0.6, 0.95
	default:
		return 1.0, 1.0, 1.0
	}
}

func modeWeights(mode string) (offense float64, defense float64, pressure float64) {
	switch mode {
	case modeOffense:
		return 1.25, 0.85, 1.0
	case modeCounterAttack:
		return 1.08, 1.05, 0.95
	default:
		return 0.85, 1.2, 1.25
	}
}

func normalizeRole(role string) string {
	r := strings.ToLower(strings.TrimSpace(role))
	switch r {
	case "goalkeeper", "defender", "fullback", "midfielder", "forward":
		return r
	default:
		return "midfielder"
	}
}

func normalizeSkill(value int) float64 {
	return clampFloat(float64(value)/100, 0, 1)
}

func isValidZone(zone string) bool {
	switch zone {
	case zoneOwnBox, zoneOwnThird, zoneMiddle, zoneFinalThird, zoneOppBox:
		return true
	default:
		return false
	}
}

func PlayMatch(home, away club.Club) match.Result {
	_ = home
	_ = away

	return match.Result{
		HomeTeamScore: 0,
		AwayTeamScore: 0,
	}
}

func normalizeTick(tick int) int {
	if tick < 1 {
		return 1
	}
	if tick > regulationTicks {
		return regulationTicks
	}
	return tick
}

func nextTick(tick int) int {
	if tick >= regulationTicks {
		return regulationTicks
	}
	return tick + 1
}

func seedForTick(seed int64, matchID uuid.UUID, tick int) int64 {
	hash := fnv.New64a()
	_, _ = hash.Write(matchID[:])

	base := int64(hash.Sum64())
	return seed + base + int64(tick*7919)
}

func goalChances(homeClubID, awayClubID uuid.UUID) (float64, float64) {
	homeSkill := float64(teamSkillFactor(homeClubID))
	awaySkill := float64(teamSkillFactor(awayClubID))

	homeChance := clampFloat(0.015+homeSkill-awaySkill/2, 0.005, 0.08)
	awayChance := clampFloat(0.012+awaySkill-homeSkill/2, 0.005, 0.08)

	return homeChance, awayChance
}

func teamSkillFactor(teamID uuid.UUID) float64 {
	hash := fnv.New64a()
	_, _ = hash.Write(teamID[:])

	raw := int(hash.Sum64()%21) - 10
	return float64(raw) / 1000
}

func clampFloat(value, min, max float64) float64 {
	return math.Min(max, math.Max(min, value))
}

func nonGoalEvent(rng *rand.Rand, tick int) (string, string) {
	frame := tacticalFrame{
		homeMode: modeOffense,
		awayMode: modeDefense,
		ballZone: zoneMiddle,
	}
	return nonGoalEventWithContext(rng, tick, frame)
}

func nonGoalEventWithContext(rng *rand.Rand, tick int, frame tacticalFrame) (string, string) {
	events := []struct {
		kind string
		desc string
	}{
		{kind: "short_pass", desc: "Passe curto para reorganizar a posse."},
		{kind: "long_pass", desc: "Lancamento buscando profundidade."},
		{kind: "dribble", desc: "Drible na faixa central."},
		{kind: "cross", desc: "Cruzamento na area e afastamento da defesa."},
		{kind: "tackle", desc: "Desarme limpo e recuperacao da bola."},
		{kind: "interception", desc: "Interceptacao de passe no meio-campo."},
		{kind: "corner", desc: "Escanteio cobrado sem conversao."},
	}

	picked := events[rng.Intn(len(events))]
	context := fmt.Sprintf("zona=%s, modo_casa=%s, modo_visitante=%s", frame.ballZone, frame.homeMode, frame.awayMode)
	return picked.kind, fmt.Sprintf("Tick %d: %s (%s)", tick, picked.desc, context)
}
