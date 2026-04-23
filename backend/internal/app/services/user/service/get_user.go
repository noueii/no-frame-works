package service

import (
	"context"

	"github.com/noueii/no-frame-works/internal/app/domain"
	"github.com/noueii/no-frame-works/internal/app/services/user"
)

// GetUser delegates to GetUserRequest.Run, which validates and fetches the
// user (and returns apperrors.NotFound when missing).
func (s *Service) GetUser(ctx context.Context, req user.GetUserRequest) (*domain.User, error) {
	return req.Run(ctx, s.repo)
}
