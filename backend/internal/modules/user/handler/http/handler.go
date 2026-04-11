package http

import "github.com/noueii/no-frame-works/internal/modules/user"

// Handler serves HTTP requests for the user module.
type Handler struct {
	api user.UserAPI
}

// NewHandler creates a new user HTTP handler.
func NewHandler(api user.UserAPI) *Handler {
	return &Handler{api: api}
}
