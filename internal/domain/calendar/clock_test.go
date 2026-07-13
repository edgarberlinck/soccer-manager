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
