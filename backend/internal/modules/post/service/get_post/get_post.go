package getpost

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/post"
)

// Execute retrieves a post by ID.
func Execute(
	ctx context.Context,
	repo post.PostRepository,
	req post.GetPostRequest,
) (post.PostView, error) {
	if err := req.Validate(); err != nil {
		return post.PostView{}, fmt.Errorf("validation failed: %w", err)
	}

	found, err := repo.FindByID(ctx, req.ID)
	if err != nil {
		return post.PostView{}, fmt.Errorf("failed to get post: %w", err)
	}

	if found == nil {
		return post.PostView{}, post.ErrPostNotFound
	}

	return post.PostView{
		ID:       found.ID,
		Title:    found.Title,
		Content:  found.Content,
		AuthorID: found.AuthorID,
	}, nil
}
