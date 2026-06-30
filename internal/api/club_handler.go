package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	repository "manager/game/internal/infrastructure/database/generated"
	"net/http"
	"strings"

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

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(clubs); err != nil {
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
