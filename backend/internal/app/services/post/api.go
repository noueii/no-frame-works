package post

import (
	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/internal/app/services/api"
)

// Service implements api.PostAPI.
type Service struct {
	app  *config.App
	repo PostRepository
}

func New(app *config.App, repo PostRepository) *Service {
	return &Service{app: app, repo: repo}
}

var _ api.PostAPI = (*Service)(nil)
