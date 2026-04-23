package handler

import (
	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/app/infrastructure/identity"
)

// Handler is the strict-server implementation. Under the god-App pattern it
// holds *config.App only; module APIs are read per-call via h.app.API().Post,
// h.app.API().User, etc.
type Handler struct {
	oapi.StrictServerInterface

	app      *config.App
	identity identity.Client
}

// NewHandler wires a new Handler. It does not construct any module services
// itself — those live on app.API(), populated by the webserver wiring step
// before this runs.
func NewHandler(app *config.App) *Handler {
	return &Handler{
		app:      app,
		identity: app.IdentityClient(),
	}
}
