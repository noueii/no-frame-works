package user

import (
	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/internal/app/services/api"
)

// Service implements api.UserAPI.
type Service struct {
	app  *config.App
	repo UserRepository
}

func New(app *config.App, repo UserRepository) *Service {
	return &Service{app: app, repo: repo}
}

var _ api.UserAPI = (*Service)(nil)
