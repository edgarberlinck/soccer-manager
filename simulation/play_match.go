package simulation

import (
	"fmt"
	"hash/fnv"
	"manager/game/internal/domain/club"
	"manager/game/internal/domain/match"
	"math"
	"math/rand"

	"github.com/google/uuid"
)

const regulationTicks = 90

type PlayMatchTickInput struct {
	MatchID     uuid.UUID
	CurrentTick int
	Seed        int64
	HomeClubID  uuid.UUID
	AwayClubID  uuid.UUID
	HomeScore   int
	AwayScore   int
}

type TickOutcome struct {
	Tick        int
	NextTick    int
	EventType   string
	Description string
	HomeScore   int
	AwayScore   int
	IsFinished  bool
}

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
		if rng.Intn(2) == 0 {
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

	homeGoalChance, awayGoalChance := goalChances(input.HomeClubID, input.AwayClubID)
	draw := rng.Float64()

	if draw < homeGoalChance {
		outcome.HomeScore++
		outcome.EventType = "goal_home"
		outcome.Description = fmt.Sprintf("Gol do mandante no tick %d. Placar: %d x %d.", tick, outcome.HomeScore, outcome.AwayScore)
	} else if draw < homeGoalChance+awayGoalChance {
		outcome.AwayScore++
		outcome.EventType = "goal_away"
		outcome.Description = fmt.Sprintf("Gol do visitante no tick %d. Placar: %d x %d.", tick, outcome.HomeScore, outcome.AwayScore)
	} else {
		outcome.EventType, outcome.Description = nonGoalEvent(rng, tick)
	}

	if tick == regulationTicks {
		outcome.EventType = "fulltime"
		outcome.Description = fmt.Sprintf("Fim de jogo: %d x %d.", outcome.HomeScore, outcome.AwayScore)
		outcome.IsFinished = true
	}

	return outcome
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

	// Faixa pequena para manter resultados realistas e previsiveis.
	return float64(hash.Sum64()%21-10) / 1000
}

func clampFloat(value, min, max float64) float64 {
	return math.Min(max, math.Max(min, value))
}

func nonGoalEvent(rng *rand.Rand, tick int) (string, string) {
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
		{kind: "foul", desc: "Falta marcada e jogo reiniciado."},
	}

	picked := events[rng.Intn(len(events))]
	return picked.kind, fmt.Sprintf("Tick %d: %s", tick, picked.desc)
}
