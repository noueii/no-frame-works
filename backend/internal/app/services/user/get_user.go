package user

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/services/api"
	"github.com/noueii/no-frame-works/internal/app/core/apperrors"
	"github.com/noueii/no-frame-works/internal/app/domain"
)

func (s *Service) GetUser(ctx context.Context, op *api.GetUserOp) (*domain.User, error) {
	if op.Request.ID == "" {
		return nil, apperrors.Validation(apperrors.CodeUserIDRequired, "user id is required", nil)
	}

	user, err := s.repo.FindByID(ctx, op.Request.ID)
	if err != nil {
		return nil, errors.Errorf("service.user.GetUser: repo find id=%s: %w", op.Request.ID, err)
	}
	if user == nil {
		return nil, apperrors.NotFound(
			apperrors.CodeUserNotFound,
			"user not found",
			map[string]any{"user_id": op.Request.ID},
		)
	}
	op.User = user
	return user, nil
}
