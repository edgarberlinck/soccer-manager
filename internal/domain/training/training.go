package training

import (
	"manager/game/internal/domain/player"
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
	case 0: return "In Progress"
	case 1: return "Finished"
	case -1: return "Cancelled"
	default: return "unknown"
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
	switch i {
	case Soft: return 0.5
	case Medium: return 1.0
	case Intense: return 1.5
	default: return 1.0
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
	// Ainda vou decidir os TrainingTypes...
	switch t {
	case Finishing: return "Finishing"
	case Passing: return "Passing"
	case Dribbling: return "Balance"
	case Shooting: return ""
	case Speed: return "Speed"
	case Strength: return ""
	case Stamina: return "Resistance & Stamina"
	case Goalkeeping: return ""
	default: return "unknown"
	}
}

type Summary struct {
	TrainingSession TrainingSession
	Score float64
}