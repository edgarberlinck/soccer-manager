package api

import (
	"encoding/json"
	repository "manager/game/internal/infrastructure/database/generated"
	"net/http"

	"github.com/google/uuid"
)

var ClubQueries *repository.Queries

func ListClubs(w http.ResponseWriter, r *http.Request) {
	if ClubQueries == nil {
		http.Error(w, "database queries not initialized", http.StatusInternalServerError)
		return
	}

	userIDParam := r.URL.Query().Get("user_id")
	if userIDParam == "" {
		http.Error(w, "missing user_id query param", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	clubs, err := ClubQueries.GetUserClubs(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to list clubs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(clubs); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}