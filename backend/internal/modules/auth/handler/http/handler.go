package http

import (
	"encoding/json"
	"net/http"

	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/config/provider"

	ory "github.com/ory/kratos-client-go"
)

// Handler handles HTTP requests for auth operations.
type Handler struct {
	kratos *ory.APIClient
	app    *config.App
}

// New creates a new auth HTTP handler.
func New(app *config.App) *Handler {
	return &Handler{
		kratos: provider.NewKratosProvider(app.EnvVars()),
		app:    app,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
