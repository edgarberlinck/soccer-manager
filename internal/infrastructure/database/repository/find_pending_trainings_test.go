package repository

import (
	"testing"
	"time"
)

func TestFindPendingTrainings(t *testing.T) {
	items := FindPendingTrainings(time.Now())
	if len(items) == 0 {
		t.Fatal("expected at least one training")
	}

	if items[0].Player.Name == "" {
		t.Fatal("expected training player name")
	}
}
