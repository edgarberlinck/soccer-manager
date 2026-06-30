package api

import (
	"encoding/json"
	"manager/game/internal/config"
	repository "manager/game/internal/infrastructure/database/generated"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(queries *repository.Queries, cfg config.Config) *chi.Mux {
	r := chi.NewRouter()

	clubHandler := NewClubHandler(queries)
	authHandler := NewAuthHandler(queries, cfg)

	r.Route("/health", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
		})
	})

	r.Route("/clubs", func(r chi.Router) {
		r.Get("/", clubHandler.ListClubs)
	})

	r.Route("/auth", func(r chi.Router) {
		r.Get("/", authHandler.HATEOAS)
		r.Post("/signup", authHandler.SignUp)
		r.Post("/signin", authHandler.SignIn)
		r.Get("/verify", authHandler.VerifyEmail)
		r.Get("/me", authHandler.Me)
	})

	return r
}
