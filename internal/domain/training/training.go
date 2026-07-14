package training

import (
	"manager/game/internal/domain/player"
	"math"
	"time"
)

type Training struct {
	Player player.Player
	StartedAt time.Time
	EndsAt time.Time
	Session TrainingSession
	// 0 for In Progress, 1 for finished and -1 for cancelled
	Status TrainingStatus
	Summary Summary
}

type TrainingStatus int

func (ts TrainingStatus) String() string {
	switch ts {
	case 0:
		return "In Progress"
	case 1:
		return "Finished"
	case -1:
		return "Cancelled"
	default:
		return "unknown"
	}
}

type TrainingSession struct {
	// Duration in minutes
	Duration time.Duration
	// Training intensity
	Intensity Intensity
	// Training type
	Type TrainingType
}

type Intensity int

const (
	Soft Intensity = iota
	Medium
	Intense
)

func (i Intensity) Multiplier() float64 {
	return i.GainMultiplier()
}

func (i Intensity) DurationMultiplier() float64 {
	switch i {
	case Soft:
		return 0.8
	case Medium:
		return 1.0
	case Intense:
		return 1.25
	default:
		return 1.0
	}
}

func (i Intensity) GainMultiplier() float64 {
	switch i {
	case Soft:
		return 0.75
	case Medium:
		return 1.0
	case Intense:
		return 1.35
	default:
		return 1.0
	}
}

func (i Intensity) FatigueMultiplier() float64 {
	switch i {
	case Soft:
		return 0.7
	case Medium:
		return 1.0
	case Intense:
		return 1.5
	default:
		return 1.0
	}
}

type TrainingType int

const (
	Finishing TrainingType = iota
	Passing
	Dribbling
	Shooting
	Speed
	Strength
	Stamina
	Goalkeeping
)

func (t TrainingType) String() string {
	switch t {
	case Finishing:
		return "Finishing"
	case Passing:
		return "Passing"
	case Dribbling:
		return "Dribbling"
	case Shooting:
		return "Shooting"
	case Speed:
		return "Speed"
	case Strength:
		return "Strength"
	case Stamina:
		return "Stamina"
	case Goalkeeping:
		return "Goalkeeping"
	default:
		return "unknown"
	}
}

func (t TrainingType) BaseDuration() time.Duration {
	switch t {
	case Finishing:
		return 60 * time.Minute
	case Passing:
		return 60 * time.Minute
	case Dribbling:
		return 55 * time.Minute
	case Shooting:
		return 75 * time.Minute
	case Speed:
		return 45 * time.Minute
	case Strength:
		return 90 * time.Minute
	case Stamina:
		return 90 * time.Minute
	case Goalkeeping:
		return 70 * time.Minute
	default:
		return 60 * time.Minute
	}
}

func (t TrainingType) BaseGain() float64 {
	switch t {
	case Speed:
		return 5.0
	case Strength, Stamina:
		return 4.0
	default:
		return 3.0
	}
}

func (t TrainingType) BaseFatigueCost() int {
	switch t {
	case Speed:
		return 8
	case Strength, Stamina:
		return 10
	case Shooting:
		return 7
	default:
		return 6
	}
}

func (t TrainingType) ApplyGain(attrs player.Attributes, gain int) player.Attributes {
	if gain <= 0 {
		return attrs
	}

	switch t {
	case Finishing:
		attrs.Finalizacao += gain
	case Passing:
		attrs.Passing += gain
	case Dribbling:
		attrs.Habilidade += gain
	case Shooting:
		attrs.Shooting += gain
	case Speed:
		attrs.Pace += gain
	case Strength:
		attrs.Fisico += gain
	case Stamina:
		attrs.FisicalStatus += gain
	case Goalkeeping:
		attrs.Impulso += gain
	}

	return attrs
}

func (s TrainingSession) ResolvedDuration() time.Duration {
	base := float64(s.Type.BaseDuration())
	resolved := time.Duration(base * s.Intensity.DurationMultiplier())
	if resolved <= 0 {
		return s.Type.BaseDuration()
	}
	return resolved
}

func ComputeTrainingGain(session TrainingSession, athlete player.Player, randomFactor float64) int {
	if randomFactor <= 0 {
		randomFactor = 1.0
	}

	base := session.Type.BaseGain()
	gain := base * session.Intensity.GainMultiplier() * ageFactor(athlete.Age) * physicalFactor(athlete.Attributes.FisicalStatus) * randomFactor
	if gain < 0 {
		return 0
	}
	if gain > 0 && gain < 1 {
		return 1
	}

	return int(math.Floor(gain))
}

func ComputeFatigueCost(session TrainingSession) int {
	base := float64(session.Type.BaseFatigueCost())
	fatigue := int(math.Round(base * session.Intensity.FatigueMultiplier()))
	if fatigue < 1 {
		return 1
	}
	return fatigue
}

func ClampPhysicalStatus(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func ageFactor(age int) float64 {
	switch {
	case age <= 23:
		return 1.15
	case age <= 29:
		return 1.0
	case age <= 34:
		return 0.85
	default:
		return 0.78
	}
}

func physicalFactor(status int) float64 {
	status = ClampPhysicalStatus(status)
	switch {
	case status >= 80:
		return 1.0
	case status >= 60:
		return 0.9
	case status >= 40:
		return 0.75
	default:
		return 0.55
	}
}

type Summary struct {
	TrainingSession TrainingSession
	Score           float64
}