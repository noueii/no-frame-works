package post

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/apperrors"
	"github.com/noueii/no-frame-works/internal/app/domain"
)

// UpdatePostRequest is the request to update a post.
type UpdatePostRequest struct {
	ID      string
	Title   string
	Content string
}

func (r UpdatePostRequest) Validate() error {
	if r.ID == "" {
		return apperrors.Validation(apperrors.CodePostIDRequired, "id is required", nil)
	}
	if r.Title == "" {
		return apperrors.Validation(apperrors.CodePostTitleRequired, "title is required", nil)
	}
	if r.Content == "" {
		return apperrors.Validation(apperrors.CodePostContentRequired, "content is required", nil)
	}
	return nil
}

// Run loads the existing post, mutates the fields the request wants to change,
// and saves the complete model back. Returns apperrors.NotFound if the post
// does not exist.
//
// Follows the fetch-mutate-save shape rather than field-specific repo methods
// so that invariants have a single enforcement site (the domain model) and
// the repo stays dumb (domain-in, domain-out).
func (r UpdatePostRequest) Run(ctx context.Context, repo PostRepository) (*domain.Post, error) {
	if err := r.Validate(); err != nil {
		return nil, errors.Errorf("post.UpdatePostRequest.Run: validate: %w", err)
	}
	existing, err := repo.FindByID(ctx, r.ID)
	if err != nil {
		return nil, errors.Errorf("post.UpdatePostRequest.Run: load existing id=%s: %w", r.ID, err)
	}
	if existing == nil {
		return nil, apperrors.NotFound(
			apperrors.CodePostNotFound,
			"post not found",
			map[string]any{"post_id": r.ID},
		)
	}
	existing.Title = r.Title
	existing.Content = r.Content
	updated, err := repo.Update(ctx, *existing)
	if err != nil {
		return nil, errors.Errorf("post.UpdatePostRequest.Run: repo update id=%s: %w", r.ID, err)
	}
	return updated, nil
}
