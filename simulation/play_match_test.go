package simulation

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"manager/game/internal/domain/club"

	"github.com/google/uuid"
)

func TestPlayMatchTickDeterministic(t *testing.T) {
	input := PlayMatchTickInput{
		MatchID:     uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		CurrentTick: 27,
		Seed:        99,
		HomeClubID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		AwayClubID:  uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		HomeScore:   1,
		AwayScore:   2,
	}

	first := PlayMatchTick(input)
	second := PlayMatchTick(input)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected deterministic output, got %+v and %+v", first, second)
	}
}

func TestPlayMatchTickSpecialTicks(t *testing.T) {
	homeID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	awayID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	matchID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	t.Run("kickoff", func(t *testing.T) {
		out := PlayMatchTick(PlayMatchTickInput{
			MatchID:     matchID,
			CurrentTick: 1,
			Seed:        1,
			HomeClubID:  homeID,
			AwayClubID:  awayID,
		})

		if out.EventType != "kickoff_home" && out.EventType != "kickoff_away" {
			t.Fatalf("unexpected kickoff event type: %s", out.EventType)
		}
		if out.NextTick != 2 {
			t.Fatalf("expected next tick 2, got %d", out.NextTick)
		}
	})

	t.Run("halftime", func(t *testing.T) {
		out := PlayMatchTick(PlayMatchTickInput{
			MatchID:     matchID,
			CurrentTick: 45,
			Seed:        1,
			HomeClubID:  homeID,
			AwayClubID:  awayID,
			HomeScore:   1,
			AwayScore:   1,
		})

		if out.EventType != "halftime" {
			t.Fatalf("expected halftime event, got %s", out.EventType)
		}
	})

	t.Run("fulltime", func(t *testing.T) {
		out := PlayMatchTick(PlayMatchTickInput{
			MatchID:     matchID,
			CurrentTick: 90,
			Seed:        1,
			HomeClubID:  homeID,
			AwayClubID:  awayID,
			HomeScore:   2,
			AwayScore:   1,
		})

		if out.EventType != "fulltime" {
			t.Fatalf("expected fulltime event, got %s", out.EventType)
		}
		if !out.IsFinished {
			t.Fatal("expected IsFinished=true at tick 90")
		}
		if out.NextTick != 90 {
			t.Fatalf("expected next tick 90, got %d", out.NextTick)
		}
	})
}

func TestPlayMatchTickBranchOutcomes(t *testing.T) {
	homeID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	awayID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	matchID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")

	find := func(kind string) TickOutcome {
		for seed := int64(1); seed <= 200000; seed++ {
			out := PlayMatchTick(PlayMatchTickInput{
				MatchID:     matchID,
				CurrentTick: 2,
				Seed:        seed,
				HomeClubID:  homeID,
				AwayClubID:  awayID,
				HomeScore:   0,
				AwayScore:   0,
			})
			if out.EventType == kind {
				return out
			}
		}

		t.Fatalf("did not find seed that produces event %s", kind)
		return TickOutcome{}
	}

	homeGoal := find("goal_home")
	if homeGoal.HomeScore != 1 || homeGoal.AwayScore != 0 {
		t.Fatalf("expected 1x0 after goal_home, got %dx%d", homeGoal.HomeScore, homeGoal.AwayScore)
	}

	awayGoal := find("goal_away")
	if awayGoal.HomeScore != 0 || awayGoal.AwayScore != 1 {
		t.Fatalf("expected 0x1 after goal_away, got %dx%d", awayGoal.HomeScore, awayGoal.AwayScore)
	}
}

func TestTickHelpers(t *testing.T) {
	if got := normalizeTick(0); got != 1 {
		t.Fatalf("expected normalized tick 1, got %d", got)
	}
	if got := normalizeTick(200); got != 90 {
		t.Fatalf("expected normalized tick 90, got %d", got)
	}
	if got := normalizeTick(11); got != 11 {
		t.Fatalf("expected normalized tick 11, got %d", got)
	}

	if got := nextTick(1); got != 2 {
		t.Fatalf("expected next tick 2, got %d", got)
	}
	if got := nextTick(90); got != 90 {
		t.Fatalf("expected next tick 90, got %d", got)
	}
}

func TestSeedAndChanceHelpers(t *testing.T) {
	matchID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	homeID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	awayID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	a := seedForTick(9, matchID, 7)
	b := seedForTick(9, matchID, 7)
	if a != b {
		t.Fatalf("expected stable seed, got %d and %d", a, b)
	}

	homeChance, awayChance := goalChances(homeID, awayID)
	if homeChance < 0.005 || homeChance > 0.08 {
		t.Fatalf("home chance out of range: %f", homeChance)
	}
	if awayChance < 0.005 || awayChance > 0.08 {
		t.Fatalf("away chance out of range: %f", awayChance)
	}

	if got := clampFloat(10, 1, 5); got != 5 {
		t.Fatalf("expected upper clamp 5, got %f", got)
	}
	if got := clampFloat(-1, 1, 5); got != 1 {
		t.Fatalf("expected lower clamp 1, got %f", got)
	}
}

func TestNonGoalEventAndPlayMatch(t *testing.T) {
	outKind, outDesc := nonGoalEvent(rand.New(rand.NewSource(123)), 12)
	if outKind == "" {
		t.Fatal("expected non-empty event kind")
	}
	if !strings.Contains(outDesc, "Tick 12") {
		t.Fatalf("unexpected event description: %s", outDesc)
	}

	result := PlayMatch(mockClub("home"), mockClub("away"))
	if result.HomeTeamScore != 0 || result.AwayTeamScore != 0 {
		t.Fatalf("expected placeholder score 0x0, got %dx%d", result.HomeTeamScore, result.AwayTeamScore)
	}
}

func mockClub(name string) club.Club {
	return club.Club{Name: name + " fc"}
}
