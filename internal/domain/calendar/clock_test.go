package calendar

import (
	"testing"
	"time"
)

func TestServerClockTickAndDayBounds(t *testing.T) {
	loc := time.FixedZone("UTC-3", -3*60*60)
	clock := NewServerClock(loc, time.Minute)

	at := time.Date(2026, time.July, 13, 10, 30, 40, 0, loc)
	start, end := clock.DayBounds(at)
	if start.Hour() != 0 || start.Minute() != 0 || start.Location() != loc {
		t.Fatalf("unexpected day start: %s", start)
	}
	if end.Sub(start) != 24*time.Hour {
		t.Fatalf("expected 24h day bounds, got %s", end.Sub(start))
	}

	startTick, endTick := clock.TickRangeForDay(at)
	if startTick > endTick {
		t.Fatalf("invalid tick range: %d > %d", startTick, endTick)
	}
	if tick := clock.TickAt(at); tick < startTick || tick > endTick {
		t.Fatalf("tick %d out of day range %d-%d", tick, startTick, endTick)
	}
}

func TestServerClockNowAndTickNow(t *testing.T) {
	loc := time.UTC
	clock := NewServerClock(loc, time.Second)
	
	now := clock.Now()
	if now.Location() != loc {
		t.Fatalf("expected location %v, got %v", loc, now.Location())
	}
	
	tickNow := clock.TickNow()
	if tickNow < 0 {
		t.Fatalf("expected positive tick, got %d", tickNow)
	}
	
	tickAt := clock.TickAt(now)
	if tickNow != tickAt {
		t.Fatalf("expected TickNow %d to equal TickAt(Now) %d", tickNow, tickAt)
	}
}

func TestServerClockTickAtConsistency(t *testing.T) {
	loc := time.UTC
	clock := NewServerClock(loc, 10*time.Second)
	
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, loc)
	
	tick1 := clock.TickAt(baseTime)
	tick2 := clock.TickAt(baseTime.Add(5 * time.Second))
	tick3 := clock.TickAt(baseTime.Add(10 * time.Second))
	tick4 := clock.TickAt(baseTime.Add(20 * time.Second))
	
	if tick1 != tick2 {
		t.Fatalf("expected same tick for times within interval: %d vs %d", tick1, tick2)
	}
	
	if tick3 <= tick2 {
		t.Fatalf("expected tick3 (%d) > tick2 (%d)", tick3, tick2)
	}
	
	if tick4 <= tick3 {
		t.Fatalf("expected tick4 (%d) > tick3 (%d)", tick4, tick3)
	}
}

func TestServerClockDayBoundsEdgeCases(t *testing.T) {
	loc := time.FixedZone("Test", 5*60*60)
	clock := NewServerClock(loc, time.Minute)
	
	tests := []struct {
		name string
		time time.Time
	}{
		{
			name: "midnight",
			time: time.Date(2026, 7, 13, 0, 0, 0, 0, loc),
		},
		{
			name: "almost midnight",
			time: time.Date(2026, 7, 13, 23, 59, 59, 999999999, loc),
		},
		{
			name: "noon",
			time: time.Date(2026, 7, 13, 12, 0, 0, 0, loc),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := clock.DayBounds(tt.time)
			
			if start.Day() != tt.time.Day() {
				t.Fatalf("start day mismatch: expected %d, got %d", tt.time.Day(), start.Day())
			}
			
			if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
				t.Fatalf("start should be at midnight: %s", start)
			}
			
			if end.Sub(start) != 24*time.Hour {
				t.Fatalf("expected 24h duration, got %s", end.Sub(start))
			}
			
			if !start.Before(end) {
				t.Fatalf("start should be before end")
			}
		})
	}
}
