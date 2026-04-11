package handler

import (
	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/infrastructure/identity"
	"github.com/noueii/no-frame-works/internal/modules/post"
	postservice "github.com/noueii/no-frame-works/internal/modules/post/service"
	"github.com/noueii/no-frame-works/internal/modules/user"
	userservice "github.com/noueii/no-frame-works/internal/modules/user/service"
	postrepo "github.com/noueii/no-frame-works/repository/post"
	userrepo "github.com/noueii/no-frame-works/repository/user"
)

type Handler struct {
	oapi.StrictServerInterface

	app      *config.App
	identity identity.Client
	postAPI  post.PostAPI
	userAPI  user.UserAPI
}

func NewHandler(app *config.App) *Handler {
	repo := postrepo.New(app.DB())
	svc := postservice.New(repo)

	idClient := app.IdentityClient()
	userRepo := userrepo.New(idClient)
	userSvc := userservice.New(userRepo)

	return &Handler{
		app:      app,
		identity: idClient,
		postAPI:  svc,
		userAPI:  userSvc,
	}
}
