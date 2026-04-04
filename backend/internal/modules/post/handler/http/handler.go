package http

import (
	"encoding/json"
	"net/http"

	"github.com/noueii/no-frame-works/internal/modules/post"
)

// Handler handles HTTP requests for the post module.
type Handler struct {
	api post.PostAPI
}

// New creates a new post HTTP handler.
func New(api post.PostAPI) *Handler {
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
