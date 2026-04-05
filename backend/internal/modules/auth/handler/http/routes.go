package http

import "github.com/go-chi/chi/v5"

// RegisterRoutes registers all auth HTTP routes on the given router.
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.Login)
		r.Post("/register", h.Register)
		r.Post("/logout", h.Logout)
		r.Get("/me", h.Me)
	})
}
