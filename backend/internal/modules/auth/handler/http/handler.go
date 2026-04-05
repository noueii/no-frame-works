package http

import (
	"encoding/json"
	"net/http"

	ory "github.com/ory/kratos-client-go"
)

// Handler handles HTTP requests for auth operations.
type Handler struct {
	kratos *ory.APIClient
}

// New creates a new auth HTTP handler.
func New(kratos *ory.APIClient) *Handler {
	return &Handler{kratos: kratos}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
