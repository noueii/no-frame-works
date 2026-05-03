package post

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/services/api"
	"github.com/noueii/no-frame-works/internal/app/core/apperrors"
	"github.com/noueii/no-frame-works/internal/app/domain"
)

func (s *Service) ListPosts(ctx context.Context, op *api.ListPostsOp) ([]domain.Post, error) {
	if op.Request.AuthorID == "" {
		return nil, apperrors.Validation(apperrors.CodePostAuthorIDRequired, "author_id is required", nil)
	}

	posts, err := s.repo.ListByAuthor(ctx, op.Request.AuthorID)
	if err != nil {
		return nil, errors.Errorf("service.post.ListPosts: repo list author=%s: %w", op.Request.AuthorID, err)
	}
	return posts, nil
}
