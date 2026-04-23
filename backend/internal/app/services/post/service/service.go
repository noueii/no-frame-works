package service

import (
	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/internal/app/services/post"
)

// Service implements post.PostAPI.
//
// Dependency shape (deliberate):
//
//   - repo: the post service's OWN repository, received directly via the
//     constructor. This is data access for post's own state. No other service
//     can reach it — it lives only as a private field on this struct.
//
//   - app:  the god-App, used exclusively for CROSS-SERVICE calls through
//     s.app.API().Other.X. It does NOT expose a Repos() accessor, so it is
//     impossible at compile time to reach another service's repository from
//     inside this service. Cross-service access is forced through the other
//     service's API interface, which is where invariants live.
//
// Each method on *Service lives in its own file named after the method
// (create_post.go, get_post.go, etc.). This file holds only the struct, the
// constructor, and the compile-time check that *Service satisfies post.PostAPI.
type Service struct {
	app  *config.App
	repo post.PostRepository
}

// New creates a new post service. The repo is injected directly (not read
// from the app) so the post service cannot accidentally reach other services'
// repositories.
func New(app *config.App, repo post.PostRepository) *Service {
	return &Service{app: app, repo: repo}
}

// Compile-time check that *Service satisfies post.PostAPI.
var _ post.PostAPI = (*Service)(nil)
