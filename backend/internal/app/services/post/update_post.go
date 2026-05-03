package post

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/services/api"
	"github.com/noueii/no-frame-works/internal/app/core/apperrors"
	"github.com/noueii/no-frame-works/internal/app/domain"
)

func (s *Service) UpdatePost(ctx context.Context, op *api.UpdatePostOp) (*domain.Post, error) {
	if op.Request.ID == "" {
		return nil, apperrors.Validation(apperrors.CodePostIDRequired, "id is required", nil)
	}
	if op.Request.Title == "" {
		return nil, apperrors.Validation(apperrors.CodePostTitleRequired, "title is required", nil)
	}
	if op.Request.Content == "" {
		return nil, apperrors.Validation(apperrors.CodePostContentRequired, "content is required", nil)
	}

	existing, err := s.repo.FindByID(ctx, op.Request.ID)
	if err != nil {
		return nil, errors.Errorf("service.post.UpdatePost: load existing id=%s: %w", op.Request.ID, err)
	}
	if existing == nil {
		return nil, apperrors.NotFound(
			apperrors.CodePostNotFound,
			"post not found",
			map[string]any{"post_id": op.Request.ID},
		)
	}

	op.Post = existing
	op.Post.Title = op.Request.Title
	op.Post.Content = op.Request.Content
	updated, err := s.repo.Update(ctx, *op.Post)
	if err != nil {
		return nil, errors.Errorf("service.post.UpdatePost: repo update id=%s: %w", op.Request.ID, err)
	}
	return updated, nil
}
