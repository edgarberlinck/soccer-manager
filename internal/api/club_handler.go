package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	repository "manager/game/internal/infrastructure/database/generated"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ClubHandler struct {
	queries *repository.Queries
}

type createClubRequest struct {
	Name         string `json:"name"`
	ShortName    string `json:"short_name"`
	Abbreviation string `json:"abbreviation"`
	Continent    string `json:"continent"`
	Country      string `json:"country"`
}

type clubSummaryResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type ensureUserClubResponse struct {
	ClubID      uuid.UUID `json:"club_id"`
	ClubName    string    `json:"club_name"`
	ClubCreated bool      `json:"club_created"`
}

type rosterPlayerResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Age             int32     `json:"age"`
	Position        string    `json:"position"`
	Overall         int16     `json:"overall"`
	Potential       int16     `json:"potential"`
	SalaryEUR       string    `json:"salary_eur"`
	ContractEndsAt  string    `json:"contract_ends_at"`
	PrimaryStrength string    `json:"primary_strength"`
}

type playerDetailResponse struct {
	ID         uuid.UUID         `json:"id"`
	Name       string            `json:"name"`
	Age        int32             `json:"age"`
	Position   string            `json:"position"`
	Overall    int16             `json:"overall"`
	Potential  int16             `json:"potential"`
	Contract   contractResponse  `json:"contract"`
	Attributes attributeResponse `json:"attributes"`
	Summary    summaryResponse   `json:"summary"`
	Matches    []matchStatItem   `json:"matches"`
}

type contractResponse struct {
	SalaryEUR        string `json:"salary_eur"`
	ReleaseClauseEUR string `json:"release_clause_eur,omitempty"`
	StartsAt         string `json:"starts_at"`
	EndsAt           string `json:"ends_at"`
}

type attributeResponse struct {
	Pace          int16 `json:"pace"`
	Passing       int16 `json:"passing"`
	Shooting      int16 `json:"shooting"`
	Altura        int16 `json:"altura"`
	Peso          int16 `json:"peso"`
	Impulso       int16 `json:"impulso"`
	Explosao      int16 `json:"explosao"`
	Fisico        int16 `json:"fisico"`
	FisicalStatus int16 `json:"fisical_status"`
	Cabeceio      int16 `json:"cabeceio"`
	Cruzamento    int16 `json:"cruzamento"`
	Habilidade    int16 `json:"habilidade"`
	Finalizacao   int16 `json:"finalizacao"`
	Dominio       int16 `json:"dominio"`
	Temperamento  int16 `json:"temperamento"`
}

type summaryResponse struct {
	Games         int64  `json:"games"`
	Goals         int64  `json:"goals"`
	Assists       int64  `json:"assists"`
	AvgRating     string `json:"avg_rating"`
	MinutesPlayed int64  `json:"minutes_played"`
}

type matchStatItem struct {
	MatchID         uuid.UUID `json:"match_id"`
	PlayedAt        string    `json:"played_at"`
	MinutesPlayed   int16     `json:"minutes_played"`
	Goals           int16     `json:"goals"`
	Assists         int16     `json:"assists"`
	Rating          string    `json:"rating"`
	PassesCompleted int16     `json:"passes_completed"`
	Shots           int16     `json:"shots"`
	Tackles         int16     `json:"tackles"`
	Saves           int16     `json:"saves"`
	Scoreline       string    `json:"scoreline"`
}

func NewClubHandler(queries *repository.Queries) *ClubHandler {
	return &ClubHandler{queries: queries}
}

func (h *ClubHandler) ListClubs(w http.ResponseWriter, r *http.Request) {
	if h.queries == nil {
		http.Error(w, "database queries not initialized", http.StatusInternalServerError)
		return
	}

	userID, ok := userIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	clubs, err := h.queries.GetUserClubs(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to list clubs", http.StatusInternalServerError)
		return
	}

	response := make([]clubSummaryResponse, 0, len(clubs))
	for _, club := range clubs {
		response = append(response, clubSummaryResponse{
			ID:   club.ID,
			Name: club.Name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *ClubHandler) CreateClub(w http.ResponseWriter, r *http.Request) {
	if h.queries == nil {
		http.Error(w, "database queries not initialized", http.StatusInternalServerError)
		return
	}

	userID, ok := userIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req createClubRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.ShortName = strings.TrimSpace(req.ShortName)
	req.Abbreviation = strings.ToUpper(strings.TrimSpace(req.Abbreviation))
	req.Continent = strings.TrimSpace(req.Continent)
	req.Country = strings.TrimSpace(req.Country)

	if req.Name == "" || req.ShortName == "" || req.Abbreviation == "" || req.Continent == "" || req.Country == "" {
		http.Error(w, "name, short_name, abbreviation, continent and country are required", http.StatusBadRequest)
		return
	}

	_, err := h.queries.GetClubByName(r.Context(), req.Name)
	if err == nil {
		http.Error(w, "club name already exists", http.StatusConflict)
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "failed to validate club name", http.StatusInternalServerError)
		return
	}

	club, err := h.queries.CreateClub(r.Context(), repository.CreateClubParams{
		ID:           uuid.New(),
		UserID:       userID,
		Name:         req.Name,
		ShortName:    sql.NullString{String: req.ShortName, Valid: true},
		Abbreviation: sql.NullString{String: req.Abbreviation, Valid: true},
		Continent:    sql.NullString{String: req.Continent, Valid: true},
		Country:      sql.NullString{String: req.Country, Valid: true},
	})
	if err != nil {
		http.Error(w, "failed to create club", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(club); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *ClubHandler) ListClubPlayers(w http.ResponseWriter, r *http.Request) {
	if h.queries == nil {
		http.Error(w, "database queries not initialized", http.StatusInternalServerError)
		return
	}

	clubID, userID, ok := h.authorizeClubRequest(w, r)
	if !ok {
		return
	}

	_, err := h.queries.GetUserClubByID(r.Context(), repository.GetUserClubByIDParams{UserID: userID, ID: clubID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "club not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load club", http.StatusInternalServerError)
		return
	}

	players, err := h.queries.ListPlayersByClubID(r.Context(), uuid.NullUUID{UUID: clubID, Valid: true})
	if err != nil {
		http.Error(w, "failed to load players", http.StatusInternalServerError)
		return
	}

	response := make([]rosterPlayerResponse, 0, len(players))
	for _, p := range players {
		response = append(response, rosterPlayerResponse{
			ID:              p.ID,
			Name:            p.Name,
			Age:             p.Age,
			Position:        p.Position,
			Overall:         p.Overall,
			Potential:       p.Potential,
			SalaryEUR:       centsToEUR(p.SalaryCents),
			ContractEndsAt:  p.EndsAt.UTC().Format(time.RFC3339),
			PrimaryStrength: pickPrimaryStrength(p.Pace, p.Passing, p.Shooting),
		})
	}

	respondJSON(w, http.StatusOK, map[string]any{"players": response})
}

func (h *ClubHandler) GetClubPlayerDetail(w http.ResponseWriter, r *http.Request) {
	if h.queries == nil {
		http.Error(w, "database queries not initialized", http.StatusInternalServerError)
		return
	}

	clubID, userID, ok := h.authorizeClubRequest(w, r)
	if !ok {
		return
	}

	_, err := h.queries.GetUserClubByID(r.Context(), repository.GetUserClubByIDParams{UserID: userID, ID: clubID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "club not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load club", http.StatusInternalServerError)
		return
	}

	playerID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "playerID")))
	if err != nil {
		http.Error(w, "invalid player id", http.StatusBadRequest)
		return
	}

	player, err := h.queries.GetPlayerByClubIDAndID(r.Context(), repository.GetPlayerByClubIDAndIDParams{
		ClubID: uuid.NullUUID{UUID: clubID, Valid: true},
		ID:     playerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "player not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load player", http.StatusInternalServerError)
		return
	}

	contract, err := h.queries.GetActivePlayerContract(r.Context(), playerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "active contract not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load contract", http.StatusInternalServerError)
		return
	}

	summary, err := h.queries.GetPlayerPerformanceSummary(r.Context(), playerID)
	if err != nil {
		http.Error(w, "failed to load summary", http.StatusInternalServerError)
		return
	}

	matchStats, err := h.queries.ListPlayerMatchStats(r.Context(), repository.ListPlayerMatchStatsParams{
		PlayerID: playerID,
		Limit:    20,
	})
	if err != nil {
		http.Error(w, "failed to load match stats", http.StatusInternalServerError)
		return
	}

	matches := make([]matchStatItem, 0, len(matchStats))
	for _, stat := range matchStats {
		playedAt := stat.CreatedAt.UTC().Format(time.RFC3339)
		if stat.FinishedAt.Valid {
			playedAt = stat.FinishedAt.Time.UTC().Format(time.RFC3339)
		}

		matches = append(matches, matchStatItem{
			MatchID:         stat.MatchID,
			PlayedAt:        playedAt,
			MinutesPlayed:   stat.MinutesPlayed,
			Goals:           stat.Goals,
			Assists:         stat.Assists,
			Rating:          stat.Rating,
			PassesCompleted: stat.PassesCompleted,
			Shots:           stat.Shots,
			Tackles:         stat.Tackles,
			Saves:           stat.Saves,
			Scoreline:       fmt.Sprintf("%d-%d", stat.HomeScore, stat.AwayScore),
		})
	}

	response := playerDetailResponse{
		ID:        player.ID,
		Name:      player.Name,
		Age:       player.Age,
		Position:  player.Position,
		Overall:   player.Overall,
		Potential: player.Potential,
		Contract: contractResponse{
			SalaryEUR:        centsToEUR(contract.SalaryCents),
			StartsAt:         contract.StartsAt.UTC().Format(time.RFC3339),
			EndsAt:           contract.EndsAt.UTC().Format(time.RFC3339),
			ReleaseClauseEUR: nullableCentsToEUR(contract.ReleaseClauseCents),
		},
		Attributes: attributeResponse{
			Pace:          player.Pace,
			Passing:       player.Passing,
			Shooting:      player.Shooting,
			Altura:        player.Altura,
			Peso:          player.Peso,
			Impulso:       player.Impulso,
			Explosao:      player.Explosao,
			Fisico:        player.Fisico,
			FisicalStatus: player.FisicalStatus,
			Cabeceio:      player.Cabeceio,
			Cruzamento:    player.Cruzamento,
			Habilidade:    player.Habilidade,
			Finalizacao:   player.Finalizacao,
			Dominio:       player.Dominio,
			Temperamento:  player.Temperamento,
		},
		Summary: summaryResponse{
			Games:         summary.Games,
			Goals:         summary.Goals,
			Assists:       summary.Assists,
			AvgRating:     summary.AvgRating,
			MinutesPlayed: summary.MinutesPlayed,
		},
		Matches: matches,
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *ClubHandler) EnsureClubSquad(w http.ResponseWriter, r *http.Request) {
	if h.queries == nil {
		http.Error(w, "database queries not initialized", http.StatusInternalServerError)
		return
	}

	clubID, userID, ok := h.authorizeClubRequest(w, r)
	if !ok {
		return
	}

	_, err := h.queries.GetUserClubByID(r.Context(), repository.GetUserClubByIDParams{UserID: userID, ID: clubID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "club not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load club", http.StatusInternalServerError)
		return
	}

	playersCount, err := h.queries.CountPlayersByClubID(r.Context(), uuid.NullUUID{UUID: clubID, Valid: true})
	if err != nil {
		http.Error(w, "failed to inspect squad", http.StatusInternalServerError)
		return
	}

	created := false
	if playersCount == 0 {
		if err := bootstrapStarterSquad(r.Context(), h.queries, clubID); err != nil {
			http.Error(w, "failed to bootstrap squad", http.StatusInternalServerError)
			return
		}
		created = true
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"club_id":          clubID,
		"squad_created":    created,
		"players_existing": playersCount,
	})
}

func (h *ClubHandler) EnsureUserClub(w http.ResponseWriter, r *http.Request) {
	if h.queries == nil {
		http.Error(w, "database queries not initialized", http.StatusInternalServerError)
		return
	}

	userID, ok := userIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to load user", http.StatusInternalServerError)
		return
	}

	club, created, err := ensureUserHasClubAndSquad(r.Context(), h.queries, userID, user.Username)
	if err != nil {
		http.Error(w, "failed to ensure club", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, ensureUserClubResponse{
		ClubID:      club.ID,
		ClubName:    club.Name,
		ClubCreated: created,
	})
}

func (h *ClubHandler) authorizeClubRequest(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return uuid.Nil, uuid.Nil, false
	}

	clubID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "clubID")))
	if err != nil {
		http.Error(w, "invalid club id", http.StatusBadRequest)
		return uuid.Nil, uuid.Nil, false
	}

	return clubID, userID, true
}

func pickPrimaryStrength(pace, passing, shooting int16) string {
	if pace >= passing && pace >= shooting {
		return "Pace"
	}
	if passing >= pace && passing >= shooting {
		return "Passing"
	}
	return "Shooting"
}

func centsToEUR(cents int64) string {
	return fmt.Sprintf("EUR %.2f", float64(cents)/100)
}

func nullableCentsToEUR(v sql.NullInt64) string {
	if !v.Valid {
		return ""
	}
	return centsToEUR(v.Int64)
}
