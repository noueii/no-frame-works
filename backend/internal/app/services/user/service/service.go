package service

import (
	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/internal/app/services/user"
)

// Service implements user.UserAPI.
//
// Like post.Service, it holds its own repository as a private field (injected
// via the constructor) and uses *config.App only for cross-service API access
// through s.app.API().Other.X. The App does not expose repos, so no other
// service can reach s.repo.
//
// Each method on *Service lives in its own file named after the method
// (get_user.go, increment_post_count.go, etc.). This file holds only the
// struct, the constructor, and the compile-time check that *Service
// satisfies user.UserAPI.
type Service struct {
	app  *config.App
	repo user.UserRepository
}

// New creates a new user service with its repository directly injected.
func New(app *config.App, repo user.UserRepository) *Service {
	return &Service{app: app, repo: repo}
}

// Compile-time check that *Service satisfies user.UserAPI.
var _ user.UserAPI = (*Service)(nil)
