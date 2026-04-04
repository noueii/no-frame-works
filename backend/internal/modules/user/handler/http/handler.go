package http

import (
	"encoding/json"
	"net/http"

	"github.com/noueii/no-frame-works/internal/modules/user"
)

// Handler handles HTTP requests for the user module.
type Handler struct {
	api user.UserAPI
}

// New creates a new user HTTP handler.
func New(api user.UserAPI) *Handler {
	return &Handler{api: api}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
