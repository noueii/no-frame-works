package post

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/apperrors"
	"github.com/noueii/no-frame-works/internal/app/domain"
)

// GetPostRequest is the request to get a post by ID.
type GetPostRequest struct {
	ID string
}

func (r GetPostRequest) Validate() error {
	if r.ID == "" {
		return apperrors.Validation(apperrors.CodePostIDRequired, "id is required", nil)
	}
	return nil
}

func (r GetPostRequest) Permission() Permission {
	return PermPostView
}

// Run validates and fetches a post by ID, returning apperrors.NotFound when
// the row is missing so handlers can map it to 404 via errors.Is.
func (r GetPostRequest) Run(ctx context.Context, repo PostRepository) (*domain.Post, error) {
	if err := r.Validate(); err != nil {
		return nil, errors.Errorf("post.GetPostRequest.Run: validate: %w", err)
	}
	found, err := repo.FindByID(ctx, r.ID)
	if err != nil {
		return nil, errors.Errorf("post.GetPostRequest.Run: repo find id=%s: %w", r.ID, err)
	}
	if found == nil {
		return nil, apperrors.NotFound(
			apperrors.CodePostNotFound,
			"post not found",
			map[string]any{"post_id": r.ID},
		)
	}
	return found, nil
}
