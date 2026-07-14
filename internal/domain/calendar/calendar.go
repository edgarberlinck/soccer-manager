package calendar

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
)

const defaultTickDuration = time.Minute

type ServerClock struct {
	Location     *time.Location
	TickDuration time.Duration
}

func NewServerClock(location *time.Location, tickDuration time.Duration) ServerClock {
	if location == nil {
		location = time.Local
	}
	if tickDuration <= 0 {
		tickDuration = defaultTickDuration
	}
	return ServerClock{Location: location, TickDuration: tickDuration}
}

func (clock ServerClock) Now() time.Time {
	return time.Now().In(clock.Location)
}

func (clock ServerClock) TickAt(at time.Time) int {
	seconds := int64(clock.TickDuration / time.Second)
	if seconds <= 0 {
		seconds = int64(defaultTickDuration / time.Second)
	}
	return int(at.Unix() / seconds)
}

func (clock ServerClock) TickNow() int {
	return clock.TickAt(clock.Now())
}

func (clock ServerClock) DayBounds(at time.Time) (time.Time, time.Time) {
	local := at.In(clock.Location)
	dayStart := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, clock.Location)
	return dayStart, dayStart.Add(24 * time.Hour)
}

func (clock ServerClock) TickRangeForDay(at time.Time) (int, int) {
	start, end := clock.DayBounds(at)
	return clock.TickAt(start), clock.TickAt(end.Add(-time.Second))
}

type EntryKind string

const (
	EntryFriendlyMatch     EntryKind = "friendly_match"
	EntryChampionshipMatch EntryKind = "championship_match"
	EntryTrainingWindow    EntryKind = "training_window"
	EntryTransferWindow    EntryKind = "transfer_window"
)

type EntryLane string

const (
	LaneSporting       EntryLane = "sporting"
	LaneAdministrative EntryLane = "administrative"
)

type CalendarEntry struct {
	ID             uuid.UUID     `json:"id"`
	ClubID         uuid.UUID     `json:"club_id"`
	Kind           EntryKind     `json:"kind"`
	Lane           EntryLane     `json:"lane"`
	Title          string        `json:"title"`
	StartTick      int           `json:"start_tick"`
	EndTick        int           `json:"end_tick"`
	MatchID        uuid.NullUUID `json:"match_id,omitempty"`
	OpponentClubID uuid.NullUUID `json:"opponent_club_id,omitempty"`
	WindowID       uuid.NullUUID `json:"window_id,omitempty"`
}

type MatchSlot struct {
	MatchID        uuid.UUID
	Kind           EntryKind
	Title          string
	StartTick      int
	EndTick        int
	OpponentClubID uuid.UUID
}

type TransferWindow struct {
	WindowID   uuid.UUID
	Title      string
	StartTick  int
	EndTick    int
}

type BuildSeasonCalendarInput struct {
	ClubID           uuid.UUID
	SeasonStartTick  int
	SeasonEndTick    int
	Matches          []MatchSlot
	TransferWindows  []TransferWindow
	TrainingTitle    string
}

type SeasonCalendar struct {
	ClubID          uuid.UUID
	SeasonStartTick int
	SeasonEndTick   int
	Sporting        []CalendarEntry
	Administrative  []CalendarEntry
}

type TickAgenda struct {
	Tick                  int
	Starting              []CalendarEntry
	Ending                []CalendarEntry
	Active                []CalendarEntry
	ActiveMatches         []CalendarEntry
	ActiveTrainingWindows []CalendarEntry
	ActiveTransferWindows []CalendarEntry
}

type SimulationBatchPlan struct {
	TotalMatches int
	BatchSize    int
	Batches      [][]CalendarEntry
}

func BuildSeasonCalendar(input BuildSeasonCalendarInput) (SeasonCalendar, error) {
	if input.ClubID == uuid.Nil {
		return SeasonCalendar{}, errors.New("club_id is required")
	}
	if input.SeasonStartTick < 1 {
		return SeasonCalendar{}, errors.New("season_start_tick must be >= 1")
	}
	if input.SeasonEndTick < input.SeasonStartTick {
		return SeasonCalendar{}, errors.New("season_end_tick must be >= season_start_tick")
	}

	sporting, err := buildSportingEntries(input)
	if err != nil {
		return SeasonCalendar{}, err
	}
	administrative, err := buildAdministrativeEntries(input)
	if err != nil {
		return SeasonCalendar{}, err
	}

	return SeasonCalendar{
		ClubID:          input.ClubID,
		SeasonStartTick: input.SeasonStartTick,
		SeasonEndTick:   input.SeasonEndTick,
		Sporting:        sporting,
		Administrative:  administrative,
	}, nil
}

func (calendar SeasonCalendar) AgendaAt(tick int) TickAgenda {
	agenda := TickAgenda{Tick: tick}
	entries := append(append([]CalendarEntry{}, calendar.Sporting...), calendar.Administrative...)
	for _, entry := range entries {
		if tick < entry.StartTick || tick > entry.EndTick {
			continue
		}
		agenda.Active = append(agenda.Active, entry)
		if tick == entry.StartTick {
			agenda.Starting = append(agenda.Starting, entry)
		}
		if tick == entry.EndTick {
			agenda.Ending = append(agenda.Ending, entry)
		}
		switch entry.Kind {
		case EntryFriendlyMatch, EntryChampionshipMatch:
			agenda.ActiveMatches = append(agenda.ActiveMatches, entry)
		case EntryTrainingWindow:
			agenda.ActiveTrainingWindows = append(agenda.ActiveTrainingWindows, entry)
		case EntryTransferWindow:
			agenda.ActiveTransferWindows = append(agenda.ActiveTransferWindows, entry)
		}
	}
	return agenda
}

func PlanMatchSimulation(entries []CalendarEntry, maxParallel, maxMatchesPerBatch int) SimulationBatchPlan {
	matches := make([]CalendarEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Kind == EntryFriendlyMatch || entry.Kind == EntryChampionshipMatch {
			matches = append(matches, entry)
		}
	}

	if maxParallel <= 0 {
		maxParallel = 1
	}
	if maxMatchesPerBatch <= 0 {
		maxMatchesPerBatch = maxParallel
	}

	batchSize := maxParallel
	if maxMatchesPerBatch < batchSize {
		batchSize = maxMatchesPerBatch
	}
	if batchSize <= 0 {
		batchSize = 1
	}

	plan := SimulationBatchPlan{TotalMatches: len(matches), BatchSize: batchSize}
	for start := 0; start < len(matches); start += batchSize {
		end := start + batchSize
		if end > len(matches) {
			end = len(matches)
		}
		batch := make([]CalendarEntry, end-start)
		copy(batch, matches[start:end])
		plan.Batches = append(plan.Batches, batch)
	}

	return plan
}

func buildSportingEntries(input BuildSeasonCalendarInput) ([]CalendarEntry, error) {
	matches := make([]MatchSlot, len(input.Matches))
	copy(matches, input.Matches)
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].StartTick == matches[j].StartTick {
			return matches[i].EndTick < matches[j].EndTick
		}
		return matches[i].StartTick < matches[j].StartTick
	})

	trainingTitle := input.TrainingTitle
	if trainingTitle == "" {
		trainingTitle = "training_window"
	}

	entries := make([]CalendarEntry, 0, len(matches)+len(matches)+1)
	nextTrainingStart := input.SeasonStartTick

	for _, match := range matches {
		if err := validateMatchSlot(match, input.SeasonStartTick, input.SeasonEndTick); err != nil {
			return nil, err
		}
		if match.StartTick < nextTrainingStart {
			return nil, fmt.Errorf("sporting overlap detected around tick %d", match.StartTick)
		}
		if nextTrainingStart < match.StartTick {
			entries = append(entries, CalendarEntry{
				ID:        uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("training:%s:%d:%d", input.ClubID, nextTrainingStart, match.StartTick-1))),
				ClubID:    input.ClubID,
				Kind:      EntryTrainingWindow,
				Lane:      LaneSporting,
				Title:     trainingTitle,
				StartTick: nextTrainingStart,
				EndTick:   match.StartTick - 1,
			})
		}

		entries = append(entries, CalendarEntry{
			ID:             match.MatchID,
			ClubID:         input.ClubID,
			Kind:           match.Kind,
			Lane:           LaneSporting,
			Title:          defaultMatchTitle(match),
			StartTick:      match.StartTick,
			EndTick:        match.EndTick,
			MatchID:        uuid.NullUUID{UUID: match.MatchID, Valid: true},
			OpponentClubID: uuid.NullUUID{UUID: match.OpponentClubID, Valid: true},
		})

		nextTrainingStart = match.EndTick + 1
	}

	if nextTrainingStart <= input.SeasonEndTick {
		entries = append(entries, CalendarEntry{
			ID:        uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("training:%s:%d:%d", input.ClubID, nextTrainingStart, input.SeasonEndTick))),
			ClubID:    input.ClubID,
			Kind:      EntryTrainingWindow,
			Lane:      LaneSporting,
			Title:     trainingTitle,
			StartTick: nextTrainingStart,
			EndTick:   input.SeasonEndTick,
		})
	}

	return entries, nil
}

func buildAdministrativeEntries(input BuildSeasonCalendarInput) ([]CalendarEntry, error) {
	entries := make([]CalendarEntry, 0, len(input.TransferWindows))
	for _, window := range input.TransferWindows {
		if window.WindowID == uuid.Nil {
			return nil, errors.New("transfer window id is required")
		}
		if window.StartTick < input.SeasonStartTick || window.EndTick > input.SeasonEndTick || window.EndTick < window.StartTick {
			return nil, fmt.Errorf("invalid transfer window %s", window.WindowID)
		}
		entries = append(entries, CalendarEntry{
			ID:        window.WindowID,
			ClubID:    input.ClubID,
			Kind:      EntryTransferWindow,
			Lane:      LaneAdministrative,
			Title:     defaultTransferTitle(window),
			StartTick: window.StartTick,
			EndTick:   window.EndTick,
			WindowID:  uuid.NullUUID{UUID: window.WindowID, Valid: true},
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].StartTick == entries[j].StartTick {
			return entries[i].EndTick < entries[j].EndTick
		}
		return entries[i].StartTick < entries[j].StartTick
	})
	return entries, nil
}

func validateMatchSlot(match MatchSlot, seasonStartTick, seasonEndTick int) error {
	if match.MatchID == uuid.Nil {
		return errors.New("match id is required")
	}
	if match.Kind != EntryFriendlyMatch && match.Kind != EntryChampionshipMatch {
		return fmt.Errorf("invalid match kind %s", match.Kind)
	}
	if match.StartTick < seasonStartTick || match.EndTick > seasonEndTick || match.EndTick < match.StartTick {
		return fmt.Errorf("invalid match window %s", match.MatchID)
	}
	if match.OpponentClubID == uuid.Nil {
		return fmt.Errorf("opponent_club_id is required for match %s", match.MatchID)
	}
	return nil
}

func defaultMatchTitle(match MatchSlot) string {
	if match.Title != "" {
		return match.Title
	}
	if match.Kind == EntryFriendlyMatch {
		return "friendly"
	}
	return "championship"
}

func defaultTransferTitle(window TransferWindow) string {
	if window.Title != "" {
		return window.Title
	}
	return "transfer_window"
}