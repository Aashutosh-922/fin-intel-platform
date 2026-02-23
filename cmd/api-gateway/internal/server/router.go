package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/handlers"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/middleware"

)

func NewRouter(h *handlers.Handler) http.Handler {
	r := chi.NewRouter()

	// Infra middleware (safe globally)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	// 🔓 Public routes (NO AUTH)
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// 🔐 Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)

		r.Route("/transactions", func(r chi.Router) {
			r.With(middleware.RequireRole("USER", "ADMIN")).
				Post("/", h.CreateTransaction)

			r.With(middleware.RequireRole("ANALYST", "ADMIN")).
				Get("/{id}", h.GetTransaction)

			r.With(middleware.RequireRole("ANALYST", "ADMIN")).
				Get("/{id}/timeline", h.GetTimeline)
		})

		r.With(middleware.RequireRole("ANALYST", "ADMIN")).
			Post("/ai/query", h.AIQuery)
	})
	
	return r
}
