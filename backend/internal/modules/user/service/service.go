package service

import (
	"context"

	"github.com/noueii/no-frame-works/internal/modules/user"
	editusername "github.com/noueii/no-frame-works/internal/modules/user/service/edit_username"
)

// Service implements user.API.
type Service struct {
	repo user.Repository
}

// New creates a new user service.
func New(repo user.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) EditUsername(
	ctx context.Context,
	req user.EditUsernameRequest,
) (*user.View, error) {
	return editusername.EditUsername(ctx, s.repo, req)
}
