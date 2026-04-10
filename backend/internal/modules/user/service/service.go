package service

import (
	"context"

	"github.com/noueii/no-frame-works/internal/modules/user"
	editusername "github.com/noueii/no-frame-works/internal/modules/user/service/edit_username"
)

// Service implements user.UserAPI.
type Service struct {
	repo user.UserRepository
}

// New creates a new user service.
func New(repo user.UserRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) EditUsername(
	ctx context.Context,
	req user.EditUsernameRequest,
) (user.UserView, error) {
	return editusername.Execute(ctx, s.repo, req)
}
