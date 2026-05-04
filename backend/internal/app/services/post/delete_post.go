package post

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/core/apperrors"
	"github.com/noueii/no-frame-works/internal/app/services/api"
)

func (s *Service) DeletePost(ctx context.Context, op *api.DeletePostOp) error {
	if op.Request.ID == "" {
		return apperrors.Validation(apperrors.CodePostIDRequired, "id is required", nil)
	}

	existing, err := s.repo.FindByID(ctx, op.Request.ID)
	if err != nil {
		return errors.Errorf("service.post.DeletePost: load existing id=%s: %w", op.Request.ID, err)
	}
	if existing == nil {
		return apperrors.NotFound(
			apperrors.CodePostNotFound,
			"post not found",
			map[string]any{"post_id": op.Request.ID},
		)
	}

	op.Post = existing
	if err := s.repo.Delete(ctx, op.Request.ID); err != nil {
		return errors.Errorf("service.post.DeletePost: repo delete id=%s: %w", op.Request.ID, err)
	}
	return nil
}
