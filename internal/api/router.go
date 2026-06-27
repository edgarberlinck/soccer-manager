package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Route("/health", func(r chi.Router) {
		r.Get("/", func (w http.ResponseWriter, r *http.Request)  {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
		})
	})
	r.Route("/clubs", func(r chi.Router) {
		r.Get("/", ListClubs)
	})

	return r
}