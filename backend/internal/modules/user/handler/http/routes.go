package http

import "github.com/go-chi/chi/v5"

// RegisterRoutes registers all user HTTP routes on the given router.
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/users", func(r chi.Router) {
		r.Post("/", h.PostCreate)
		r.Get("/{id}", h.GetUser)
	})
}
