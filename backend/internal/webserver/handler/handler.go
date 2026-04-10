package handler

import (
	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/infrastructure/identity"
	"github.com/noueii/no-frame-works/internal/modules/post"
	postmw "github.com/noueii/no-frame-works/internal/modules/post/middleware"
	postservice "github.com/noueii/no-frame-works/internal/modules/post/service"
	postrepo "github.com/noueii/no-frame-works/repository/post"
)

type Handler struct {
	oapi.StrictServerInterface

	app      *config.App
	identity identity.Client
	postAPI  post.PostAPI
}

func NewHandler(app *config.App) *Handler {
	repo := postrepo.New(app.DB())
	svc := postservice.New(repo)
	api := postmw.NewPermissionLayer(svc, repo)

	return &Handler{
		app:      app,
		identity: app.IdentityClient(),
		postAPI:  api,
	}
}
