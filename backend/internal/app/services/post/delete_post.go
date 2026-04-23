package post

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/apperrors"
)

// DeletePostRequest is the request to delete a post.
type DeletePostRequest struct {
	ID string
}

func (r DeletePostRequest) Validate() error {
	if r.ID == "" {
		return apperrors.Validation(apperrors.CodePostIDRequired, "id is required", nil)
	}
	return nil
}

// Run validates, ensures the post exists, and removes it. The pre-delete
// FindByID is intentional: it turns "delete of nonexistent post" into a 404
// rather than a silent success, which matches what most clients expect.
func (r DeletePostRequest) Run(ctx context.Context, repo PostRepository) error {
	if err := r.Validate(); err != nil {
		return errors.Errorf("post.DeletePostRequest.Run: validate: %w", err)
	}
	existing, err := repo.FindByID(ctx, r.ID)
	if err != nil {
		return errors.Errorf("post.DeletePostRequest.Run: load existing id=%s: %w", r.ID, err)
	}
	if existing == nil {
		return apperrors.NotFound(
			apperrors.CodePostNotFound,
			"post not found",
			map[string]any{"post_id": r.ID},
		)
	}
	if err := repo.Delete(ctx, r.ID); err != nil {
		return errors.Errorf("post.DeletePostRequest.Run: repo delete id=%s: %w", r.ID, err)
	}
	return nil
}
