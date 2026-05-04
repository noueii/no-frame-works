package post

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/core/apperrors"
	"github.com/noueii/no-frame-works/internal/app/domain"
	"github.com/noueii/no-frame-works/internal/app/services/api"
)

func (s *Service) CreatePost(ctx context.Context, op *api.CreatePostOp) (*domain.Post, error) {
	if op.Request.Title == "" {
		return nil, apperrors.Validation(apperrors.CodePostTitleRequired, "title is required", nil)
	}
	if op.Request.Content == "" {
		return nil, apperrors.Validation(
			apperrors.CodePostContentRequired,
			"content is required",
			nil,
		)
	}
	if op.Request.AuthorID == "" {
		return nil, apperrors.Validation(
			apperrors.CodePostAuthorIDRequired,
			"author_id is required",
			nil,
		)
	}

	created, err := s.repo.Create(ctx, domain.Post{
		Title:    op.Request.Title,
		Content:  op.Request.Content,
		AuthorID: op.Request.AuthorID,
	})
	if err != nil {
		return nil, errors.Errorf("service.post.CreatePost: repo create: %w", err)
	}

	if err := s.app.API().Users.IncrementPostCount(ctx, &api.IncrementPostCountOp{
		Request: api.IncrementPostCountRequest{UserID: op.Request.AuthorID},
	}); err != nil {
		return nil, errors.Errorf(
			"service.post.CreatePost: increment author=%s: %w",
			op.Request.AuthorID,
			err,
		)
	}

	return created, nil
}
