package user

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/services/api"
	"github.com/noueii/no-frame-works/internal/app/core/apperrors"
)

func (s *Service) IncrementPostCount(ctx context.Context, op *api.IncrementPostCountOp) error {
	if op.Request.UserID == "" {
		return apperrors.Validation(apperrors.CodeUserIDRequired, "user id is required", nil)
	}

	if err := s.repo.IncrementPostCount(ctx, op.Request.UserID); err != nil {
		return errors.Errorf("service.user.IncrementPostCount: repo increment id=%s: %w", op.Request.UserID, err)
	}
	return nil
}
