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
	r.Use(NewCORSMiddleware(cfg))
	authMiddleware := NewJWTAuthMiddleware(cfg)

	clubHandler := NewClubHandler(queries)
	authHandler := NewAuthHandler(queries, cfg)
	calendarHandler := NewCalendarHandler(queries)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)

		r.Route("/health", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
			})
		})

		r.Route("/clubs", func(r chi.Router) {
			r.Get("/", clubHandler.ListClubs)
			r.Post("/ensure", clubHandler.EnsureUserClub)
			r.Post("/", clubHandler.CreateClub)
			r.Post("/{clubID}/ensure-squad", clubHandler.EnsureClubSquad)
			r.Get("/{clubID}/players", clubHandler.ListClubPlayers)
			r.Get("/{clubID}/players/{playerID}", clubHandler.GetClubPlayerDetail)
		})

		r.Route("/calendar", func(r chi.Router) {
			r.Get("/season", calendarHandler.GetSeasonCalendar)
			r.Get("/clubs/{clubId}", calendarHandler.GetClubCalendar)
		})
	})

	r.Route("/auth", func(r chi.Router) {
		r.Get("/", authHandler.HATEOAS)
		r.Post("/signup", authHandler.SignUp)
		r.Post("/signin", authHandler.SignIn)
		r.Get("/verify", authHandler.VerifyEmail)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Get("/me", authHandler.Me)
		})
	})

	return r
}
