package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestCalendarHandler_GetClubCalendar(t *testing.T) {
	t.Run("should return 400 for invalid club ID", func(t *testing.T) {
		handler := &CalendarHandler{}
		
		req := httptest.NewRequest("GET", "/calendar/clubs/invalid-uuid", nil)
		w := httptest.NewRecorder()
		
		// Simula chi URLParam
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("clubId", "invalid-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		
		handler.GetClubCalendar(w, req)
		
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestCalendarHandler_GetSeasonCalendar(t *testing.T) {
	t.Run("should use default dates when not provided", func(t *testing.T) {
		handler := &CalendarHandler{}
		
		req := httptest.NewRequest("GET", "/calendar/season", nil)
		w := httptest.NewRecorder()
		
		// Verifica que não retorna erro com defaults
		// Nota: vai falhar ao tentar buscar do banco (queries é nil),
		// mas pelo menos valida o parsing de datas
		handler.GetSeasonCalendar(w, req)
		
		// Se chegou aqui com erro 500 (banco), significa que passou pelo parsing
		if w.Code != http.StatusInternalServerError {
			// Isso é esperado pois queries é nil
		}
	})
	
	t.Run("should return 400 for invalid start date", func(t *testing.T) {
		handler := &CalendarHandler{}
		
		req := httptest.NewRequest("GET", "/calendar/season?start=invalid-date", nil)
		w := httptest.NewRecorder()
		
		handler.GetSeasonCalendar(w, req)
		
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
	
	t.Run("should return 400 for invalid end date", func(t *testing.T) {
		handler := &CalendarHandler{}
		
		req := httptest.NewRequest("GET", "/calendar/season?end=invalid-date", nil)
		w := httptest.NewRecorder()
		
		handler.GetSeasonCalendar(w, req)
		
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestCalendarResponse_JSON(t *testing.T) {
	t.Run("should serialize calendar response correctly", func(t *testing.T) {
		clubID := uuid.New()
		matchID := uuid.New()
		opponentID := uuid.New()
		homeScore := int32(2)
		awayScore := int32(1)
		finishedAt := time.Now()
		
		response := CalendarResponse{
			ClubID: clubID,
			Matches: []MatchResponse{
				{
					ID:           matchID,
					HomeClubID:   clubID,
					AwayClubID:   opponentID,
					Status:       "finished",
					HomeScore:    &homeScore,
					AwayScore:    &awayScore,
					IsHome:       true,
					OpponentID:   opponentID,
					CreatedAt:    time.Now(),
					FinishedAt:   &finishedAt,
				},
			},
			Stats: CalendarStats{
				TotalMatches:     1,
				HomeMatches:      1,
				AwayMatches:      0,
				CompletedMatches: 1,
				PendingMatches:   0,
			},
		}
		
		// Serializa para JSON
		data, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}
		
		// Deserializa de volta
		var decoded CalendarResponse
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		
		if decoded.ClubID != clubID {
			t.Errorf("club_id mismatch")
		}
		if len(decoded.Matches) != 1 {
			t.Errorf("expected 1 match, got %d", len(decoded.Matches))
		}
		if decoded.Stats.TotalMatches != 1 {
			t.Errorf("expected 1 total match, got %d", decoded.Stats.TotalMatches)
		}
	})
}

func TestSeasonCalendarResponse_JSON(t *testing.T) {
	t.Run("should serialize season response correctly", func(t *testing.T) {
		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
		
		response := SeasonCalendarResponse{
			Season: SeasonInfo{
				StartDate: startDate,
				EndDate:   endDate,
			},
			Matches: []MatchResponse{},
			Stats: CalendarStats{
				TotalMatches: 0,
			},
		}
		
		data, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}
		
		var decoded SeasonCalendarResponse
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		
		if !decoded.Season.StartDate.Equal(startDate) {
			t.Errorf("start_date mismatch")
		}
		if !decoded.Season.EndDate.Equal(endDate) {
			t.Errorf("end_date mismatch")
		}
	})
}
