package calendar

import (
	"testing"

	"github.com/google/uuid"
)

func TestBuildSeasonCalendarFillsTrainingGaps(t *testing.T) {
	clubID := uuid.New()
	opp1 := uuid.New()
	opp2 := uuid.New()
	calendar, err := BuildSeasonCalendar(BuildSeasonCalendarInput{
		ClubID:          clubID,
		SeasonStartTick: 1,
		SeasonEndTick:   20,
		Matches: []MatchSlot{
			{MatchID: uuid.New(), Kind: EntryFriendlyMatch, StartTick: 5, EndTick: 6, OpponentClubID: opp1},
			{MatchID: uuid.New(), Kind: EntryChampionshipMatch, StartTick: 10, EndTick: 12, OpponentClubID: opp2},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(calendar.Sporting) != 5 {
		t.Fatalf("expected 5 sporting entries, got %d", len(calendar.Sporting))
	}

	assertWindow(t, calendar.Sporting[0], EntryTrainingWindow, 1, 4)
	assertWindow(t, calendar.Sporting[1], EntryFriendlyMatch, 5, 6)
	assertWindow(t, calendar.Sporting[2], EntryTrainingWindow, 7, 9)
	assertWindow(t, calendar.Sporting[3], EntryChampionshipMatch, 10, 12)
	assertWindow(t, calendar.Sporting[4], EntryTrainingWindow, 13, 20)
}

func TestAgendaAtIncludesTransferAndMatchLanes(t *testing.T) {
	clubID := uuid.New()
	opp := uuid.New()
	windowID := uuid.New()
	calendar, err := BuildSeasonCalendar(BuildSeasonCalendarInput{
		ClubID:          clubID,
		SeasonStartTick: 1,
		SeasonEndTick:   15,
		Matches: []MatchSlot{
			{MatchID: uuid.New(), Kind: EntryChampionshipMatch, StartTick: 8, EndTick: 10, OpponentClubID: opp},
		},
		TransferWindows: []TransferWindow{{WindowID: windowID, StartTick: 7, EndTick: 12}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	agenda := calendar.AgendaAt(8)
	if len(agenda.ActiveMatches) != 1 {
		t.Fatalf("expected 1 active match, got %d", len(agenda.ActiveMatches))
	}
	if len(agenda.ActiveTransferWindows) != 1 {
		t.Fatalf("expected 1 active transfer window, got %d", len(agenda.ActiveTransferWindows))
	}
	if len(agenda.Starting) != 1 {
		t.Fatalf("expected match to start at tick 8")
	}
}

func TestPlanMatchSimulationRespectsLimits(t *testing.T) {
	clubID := uuid.New()
	entries := []CalendarEntry{
		{ID: uuid.New(), ClubID: clubID, Kind: EntryFriendlyMatch},
		{ID: uuid.New(), ClubID: clubID, Kind: EntryChampionshipMatch},
		{ID: uuid.New(), ClubID: clubID, Kind: EntryChampionshipMatch},
		{ID: uuid.New(), ClubID: clubID, Kind: EntryChampionshipMatch},
		{ID: uuid.New(), ClubID: clubID, Kind: EntryTrainingWindow},
	}

	plan := PlanMatchSimulation(entries, 4, 2)
	if plan.TotalMatches != 4 {
		t.Fatalf("expected 4 matches, got %d", plan.TotalMatches)
	}
	if plan.BatchSize != 2 {
		t.Fatalf("expected batch size 2, got %d", plan.BatchSize)
	}
	if len(plan.Batches) != 2 {
		t.Fatalf("expected 2 batches, got %d", len(plan.Batches))
	}
	for _, batch := range plan.Batches {
		if len(batch) > 2 {
			t.Fatalf("batch exceeds cap: %d", len(batch))
		}
	}
}

func TestBuildSeasonCalendarRejectsOverlappingMatches(t *testing.T) {
	_, err := BuildSeasonCalendar(BuildSeasonCalendarInput{
		ClubID:          uuid.New(),
		SeasonStartTick: 1,
		SeasonEndTick:   10,
		Matches: []MatchSlot{
			{MatchID: uuid.New(), Kind: EntryFriendlyMatch, StartTick: 3, EndTick: 5, OpponentClubID: uuid.New()},
			{MatchID: uuid.New(), Kind: EntryChampionshipMatch, StartTick: 5, EndTick: 6, OpponentClubID: uuid.New()},
		},
	})
	if err == nil {
		t.Fatal("expected overlap validation error")
	}
}

func assertWindow(t *testing.T, entry CalendarEntry, kind EntryKind, start, end int) {
	t.Helper()
	if entry.Kind != kind || entry.StartTick != start || entry.EndTick != end {
		t.Fatalf("unexpected entry: %+v, want kind=%s %d-%d", entry, kind, start, end)
	}
}

func TestBuildSeasonCalendarWithTransferWindows(t *testing.T) {
	clubID := uuid.New()
	windowID := uuid.New()
	
	tests := []struct {
		name        string
		input       BuildSeasonCalendarInput
		expectError bool
	}{
		{
			name: "valid transfer window with custom title",
			input: BuildSeasonCalendarInput{
				ClubID:          clubID,
				SeasonStartTick: 1,
				SeasonEndTick:   100,
				TransferWindows: []TransferWindow{
					{WindowID: windowID, StartTick: 10, EndTick: 20, Title: "Summer Transfer"},
				},
			},
			expectError: false,
		},
		{
			name: "valid transfer window without title",
			input: BuildSeasonCalendarInput{
				ClubID:          clubID,
				SeasonStartTick: 1,
				SeasonEndTick:   100,
				TransferWindows: []TransferWindow{
					{WindowID: windowID, StartTick: 10, EndTick: 20},
				},
			},
			expectError: false,
		},
		{
			name: "transfer window with nil ID",
			input: BuildSeasonCalendarInput{
				ClubID:          clubID,
				SeasonStartTick: 1,
				SeasonEndTick:   100,
				TransferWindows: []TransferWindow{
					{WindowID: uuid.Nil, StartTick: 10, EndTick: 20},
				},
			},
			expectError: true,
		},
		{
			name: "transfer window before season start",
			input: BuildSeasonCalendarInput{
				ClubID:          clubID,
				SeasonStartTick: 10,
				SeasonEndTick:   100,
				TransferWindows: []TransferWindow{
					{WindowID: windowID, StartTick: 5, EndTick: 20},
				},
			},
			expectError: true,
		},
		{
			name: "transfer window after season end",
			input: BuildSeasonCalendarInput{
				ClubID:          clubID,
				SeasonStartTick: 1,
				SeasonEndTick:   50,
				TransferWindows: []TransferWindow{
					{WindowID: windowID, StartTick: 10, EndTick: 60},
				},
			},
			expectError: true,
		},
		{
			name: "transfer window with end before start",
			input: BuildSeasonCalendarInput{
				ClubID:          clubID,
				SeasonStartTick: 1,
				SeasonEndTick:   100,
				TransferWindows: []TransferWindow{
					{WindowID: windowID, StartTick: 30, EndTick: 20},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calendar, err := BuildSeasonCalendar(tt.input)
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(tt.input.TransferWindows) > 0 {
					if len(calendar.Administrative) != len(tt.input.TransferWindows) {
						t.Fatalf("expected %d administrative entries, got %d", len(tt.input.TransferWindows), len(calendar.Administrative))
					}
				}
			}
		})
	}
}

func TestBuildSeasonCalendarValidatesMatches(t *testing.T) {
	clubID := uuid.New()
	opponentID := uuid.New()

	tests := []struct {
		name        string
		input       BuildSeasonCalendarInput
		expectError bool
	}{
		{
			name: "match with nil ID",
			input: BuildSeasonCalendarInput{
				ClubID:          clubID,
				SeasonStartTick: 1,
				SeasonEndTick:   100,
				Matches: []MatchSlot{
					{MatchID: uuid.Nil, Kind: EntryFriendlyMatch, StartTick: 10, EndTick: 12, OpponentClubID: opponentID},
				},
			},
			expectError: true,
		},
		{
			name: "match with invalid kind",
			input: BuildSeasonCalendarInput{
				ClubID:          clubID,
				SeasonStartTick: 1,
				SeasonEndTick:   100,
				Matches: []MatchSlot{
					{MatchID: uuid.New(), Kind: "invalid_kind", StartTick: 10, EndTick: 12, OpponentClubID: opponentID},
				},
			},
			expectError: true,
		},
		{
			name: "match before season start",
			input: BuildSeasonCalendarInput{
				ClubID:          clubID,
				SeasonStartTick: 10,
				SeasonEndTick:   100,
				Matches: []MatchSlot{
					{MatchID: uuid.New(), Kind: EntryFriendlyMatch, StartTick: 5, EndTick: 12, OpponentClubID: opponentID},
				},
			},
			expectError: true,
		},
		{
			name: "match after season end",
			input: BuildSeasonCalendarInput{
				ClubID:          clubID,
				SeasonStartTick: 1,
				SeasonEndTick:   50,
				Matches: []MatchSlot{
					{MatchID: uuid.New(), Kind: EntryFriendlyMatch, StartTick: 40, EndTick: 60, OpponentClubID: opponentID},
				},
			},
			expectError: true,
		},
		{
			name: "match with end before start",
			input: BuildSeasonCalendarInput{
				ClubID:          clubID,
				SeasonStartTick: 1,
				SeasonEndTick:   100,
				Matches: []MatchSlot{
					{MatchID: uuid.New(), Kind: EntryFriendlyMatch, StartTick: 30, EndTick: 20, OpponentClubID: opponentID},
				},
			},
			expectError: true,
		},
		{
			name: "match without opponent",
			input: BuildSeasonCalendarInput{
				ClubID:          clubID,
				SeasonStartTick: 1,
				SeasonEndTick:   100,
				Matches: []MatchSlot{
					{MatchID: uuid.New(), Kind: EntryFriendlyMatch, StartTick: 10, EndTick: 12, OpponentClubID: uuid.Nil},
				},
			},
			expectError: true,
		},
		{
			name: "valid friendly match with custom title",
			input: BuildSeasonCalendarInput{
				ClubID:          clubID,
				SeasonStartTick: 1,
				SeasonEndTick:   100,
				Matches: []MatchSlot{
					{MatchID: uuid.New(), Kind: EntryFriendlyMatch, StartTick: 10, EndTick: 12, OpponentClubID: opponentID, Title: "Test Match"},
				},
			},
			expectError: false,
		},
		{
			name: "valid championship match without title",
			input: BuildSeasonCalendarInput{
				ClubID:          clubID,
				SeasonStartTick: 1,
				SeasonEndTick:   100,
				Matches: []MatchSlot{
					{MatchID: uuid.New(), Kind: EntryChampionshipMatch, StartTick: 10, EndTick: 12, OpponentClubID: opponentID},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildSeasonCalendar(tt.input)
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAgendaAtEdgeCases(t *testing.T) {
	clubID := uuid.New()
	opponentID := uuid.New()
	
	calendar, err := BuildSeasonCalendar(BuildSeasonCalendarInput{
		ClubID:          clubID,
		SeasonStartTick: 1,
		SeasonEndTick:   100,
		Matches: []MatchSlot{
			{MatchID: uuid.New(), Kind: EntryFriendlyMatch, StartTick: 10, EndTick: 15, OpponentClubID: opponentID},
			{MatchID: uuid.New(), Kind: EntryChampionshipMatch, StartTick: 20, EndTick: 25, OpponentClubID: opponentID},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("before any match", func(t *testing.T) {
		agenda := calendar.AgendaAt(5)
		if len(agenda.ActiveMatches) != 0 {
			t.Fatalf("expected no active matches, got %d", len(agenda.ActiveMatches))
		}
		if len(agenda.Starting) != 0 {
			t.Fatalf("expected no starting matches, got %d", len(agenda.Starting))
		}
	})

	t.Run("at match start", func(t *testing.T) {
		agenda := calendar.AgendaAt(10)
		if len(agenda.Starting) != 1 {
			t.Fatalf("expected 1 starting match, got %d", len(agenda.Starting))
		}
	})

	t.Run("during match", func(t *testing.T) {
		agenda := calendar.AgendaAt(12)
		if len(agenda.ActiveMatches) != 1 {
			t.Fatalf("expected 1 active match, got %d", len(agenda.ActiveMatches))
		}
	})

	t.Run("after all matches", func(t *testing.T) {
		agenda := calendar.AgendaAt(50)
		if len(agenda.ActiveMatches) != 0 {
			t.Fatalf("expected no active matches, got %d", len(agenda.ActiveMatches))
		}
	})
}