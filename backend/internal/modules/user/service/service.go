package service

import (
	"context"

	"github.com/noueii/no-frame-works/internal/modules/user"
	createuser "github.com/noueii/no-frame-works/internal/modules/user/service/create_user"
	getuser "github.com/noueii/no-frame-works/internal/modules/user/service/get_user"
)

// Service implements user.UserAPI.
type Service struct {
	repo user.UserRepository
}

// New creates a new user service.
func New(repo user.UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(
	ctx context.Context,
	req user.CreateUserRequest,
) (user.UserView, error) {
	return createuser.Execute(ctx, s.repo, req)
}

func (s *Service) GetUser(ctx context.Context, req user.GetUserRequest) (user.UserView, error) {
	return getuser.Execute(ctx, s.repo, req)
}
