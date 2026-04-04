package http

import "github.com/go-chi/chi/v5"

// RegisterRoutes registers all post HTTP routes on the given router.
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/posts", func(r chi.Router) {
		r.Post("/", h.PostCreate)
		r.Get("/{id}", h.GetPost)
		r.Get("/by-author/{authorId}", h.ListPosts)
	})
}
