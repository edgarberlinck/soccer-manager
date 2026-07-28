package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	repository "manager/game/internal/infrastructure/database/generated"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CalendarHandler struct {
	queries *repository.Queries
}

func NewCalendarHandler(queries *repository.Queries) *CalendarHandler {
	return &CalendarHandler{queries: queries}
}

type CalendarResponse struct {
	ClubID  uuid.UUID       `json:"club_id"`
	Matches []MatchResponse `json:"matches"`
	Stats   CalendarStats   `json:"stats"`
}

type MatchResponse struct {
	ID           uuid.UUID  `json:"id"`
	HomeClubID   uuid.UUID  `json:"home_club_id"`
	AwayClubID   uuid.UUID  `json:"away_club_id"`
	Status       string     `json:"status"`
	HomeScore    *int32     `json:"home_score,omitempty"`
	AwayScore    *int32     `json:"away_score,omitempty"`
	IsHome       bool       `json:"is_home"`
	OpponentID   uuid.UUID  `json:"opponent_id"`
	CreatedAt    time.Time  `json:"created_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
}

type CalendarStats struct {
	TotalMatches    int `json:"total_matches"`
	HomeMatches     int `json:"home_matches"`
	AwayMatches     int `json:"away_matches"`
	CompletedMatches int `json:"completed_matches"`
	PendingMatches  int `json:"pending_matches"`
}

// GetClubCalendar retorna o calendário de um clube
func (h *CalendarHandler) GetClubCalendar(w http.ResponseWriter, r *http.Request) {
	clubIDStr := chi.URLParam(r, "clubId")
	clubID, err := uuid.Parse(clubIDStr)
	if err != nil {
		http.Error(w, "Invalid club ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	matches, err := h.queries.GetClubMatches(ctx, clubID)
	if err != nil {
		log.Printf("Error fetching club matches: %v", err)
		http.Error(w, "Error fetching matches", http.StatusInternalServerError)
		return
	}

	// Converte para response
	matchResponses := make([]MatchResponse, 0, len(matches))
	stats := CalendarStats{}

	for _, match := range matches {
		isHome := match.HomeClubID == clubID
		opponentID := match.AwayClubID
		if !isHome {
			opponentID = match.HomeClubID
		}

		matchResp := MatchResponse{
			ID:         match.ID,
			HomeClubID: match.HomeClubID,
			AwayClubID: match.AwayClubID,
			Status:     match.Status,
			IsHome:     isHome,
			OpponentID: opponentID,
			CreatedAt:  match.CreatedAt,
		}

		if match.HomeScore.Valid {
			score := int32(match.HomeScore.Int32)
			matchResp.HomeScore = &score
		}
		if match.AwayScore.Valid {
			score := int32(match.AwayScore.Int32)
			matchResp.AwayScore = &score
		}
		if match.FinishedAt.Valid {
			matchResp.FinishedAt = &match.FinishedAt.Time
		}

		matchResponses = append(matchResponses, matchResp)

		// Atualiza estatísticas
		stats.TotalMatches++
		if isHome {
			stats.HomeMatches++
		} else {
			stats.AwayMatches++
		}
		if match.Status == "finished" {
			stats.CompletedMatches++
		} else {
			stats.PendingMatches++
		}
	}

	response := CalendarResponse{
		ClubID:  clubID,
		Matches: matchResponses,
		Stats:   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type SeasonCalendarResponse struct {
	Season  SeasonInfo      `json:"season"`
	Matches []MatchResponse `json:"matches"`
	Stats   CalendarStats   `json:"stats"`
}

type SeasonInfo struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// GetSeasonCalendar retorna o calendário da temporada
func (h *CalendarHandler) GetSeasonCalendar(w http.ResponseWriter, r *http.Request) {
	// Busca parâmetros de query
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var startDate, endDate time.Time
	var err error

	if startStr != "" {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			http.Error(w, "Invalid start date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		// Default: início do ano atual
		now := time.Now()
		startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	}

	if endStr != "" {
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			http.Error(w, "Invalid end date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		// Default: fim do ano atual
		now := time.Now()
		endDate = time.Date(now.Year(), 12, 31, 23, 59, 59, 0, time.UTC)
	}

	ctx := r.Context()
	matches, err := h.queries.GetSeasonMatches(ctx, repository.GetSeasonMatchesParams{
		CreatedAt:   startDate,
		CreatedAt_2: endDate,
	})
	if err != nil {
		log.Printf("Error fetching season matches: %v", err)
		http.Error(w, "Error fetching matches", http.StatusInternalServerError)
		return
	}

	// Converte para response
	matchResponses := make([]MatchResponse, 0, len(matches))
	stats := CalendarStats{}

	for _, match := range matches {
		matchResp := MatchResponse{
			ID:         match.ID,
			HomeClubID: match.HomeClubID,
			AwayClubID: match.AwayClubID,
			Status:     match.Status,
			IsHome:     false, // Não aplicável na view geral
			OpponentID: uuid.Nil,
			CreatedAt:  match.CreatedAt,
		}

		if match.HomeScore.Valid {
			score := int32(match.HomeScore.Int32)
			matchResp.HomeScore = &score
		}
		if match.AwayScore.Valid {
			score := int32(match.AwayScore.Int32)
			matchResp.AwayScore = &score
		}
		if match.FinishedAt.Valid {
			matchResp.FinishedAt = &match.FinishedAt.Time
		}

		matchResponses = append(matchResponses, matchResp)

		stats.TotalMatches++
		if match.Status == "finished" {
			stats.CompletedMatches++
		} else {
			stats.PendingMatches++
		}
	}

	response := SeasonCalendarResponse{
		Season: SeasonInfo{
			StartDate: startDate,
			EndDate:   endDate,
		},
		Matches: matchResponses,
		Stats:   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
