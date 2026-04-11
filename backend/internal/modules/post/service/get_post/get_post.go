package getpost

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/modules/post"
)

// GetPost retrieves a post by ID.
func GetPost(
	ctx context.Context,
	repo post.Repository,
	req post.GetPostRequest,
) (*post.View, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if err := req.CheckPermission(ctx); err != nil {
		return nil, err
	}

	found, err := repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, errors.Errorf("failed to get post: %w", err)
	}

	if found == nil {
		return nil, post.ErrPostNotFound
	}

	return &post.View{
		ID:       found.ID,
		Title:    found.Title,
		Content:  found.Content,
		AuthorID: found.AuthorID,
	}, nil
}
