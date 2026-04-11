package http

import "github.com/go-chi/chi/v5"

// Routes registers the user module's HTTP routes.
func (h *Handler) Routes(r chi.Router) {
	r.Put("/users/{id}/username", h.editUsername)
}
