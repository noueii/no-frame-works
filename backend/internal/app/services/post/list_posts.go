package post

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/apperrors"
	"github.com/noueii/no-frame-works/internal/app/domain"
)

// ListPostsRequest is the request to list posts by author.
type ListPostsRequest struct {
	AuthorID string
}

func (r ListPostsRequest) Validate() error {
	if r.AuthorID == "" {
		return apperrors.Validation(apperrors.CodePostAuthorIDRequired, "author_id is required", nil)
	}
	return nil
}

func (r ListPostsRequest) Permission() Permission {
	return PermPostList
}

// Run validates and returns every post for the given author.
func (r ListPostsRequest) Run(ctx context.Context, repo PostRepository) ([]domain.Post, error) {
	if err := r.Validate(); err != nil {
		return nil, errors.Errorf("post.ListPostsRequest.Run: validate: %w", err)
	}
	posts, err := repo.ListByAuthor(ctx, r.AuthorID)
	if err != nil {
		return nil, errors.Errorf("post.ListPostsRequest.Run: repo list author=%s: %w", r.AuthorID, err)
	}
	return posts, nil
}
