package post

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/services/api"
	"github.com/noueii/no-frame-works/internal/app/core/apperrors"
	"github.com/noueii/no-frame-works/internal/app/domain"
)

func (s *Service) GetPost(ctx context.Context, op *api.GetPostOp) (*domain.Post, error) {
	if op.Request.ID == "" {
		return nil, apperrors.Validation(apperrors.CodePostIDRequired, "id is required", nil)
	}

	post, err := s.repo.FindByID(ctx, op.Request.ID)
	if err != nil {
		return nil, errors.Errorf("service.post.GetPost: repo find id=%s: %w", op.Request.ID, err)
	}
	if post == nil {
		return nil, apperrors.NotFound(
			apperrors.CodePostNotFound,
			"post not found",
			map[string]any{"post_id": op.Request.ID},
		)
	}
	op.Post = post
	return post, nil
}
