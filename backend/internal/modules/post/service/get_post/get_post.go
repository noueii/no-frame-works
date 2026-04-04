package getpost

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/post"
	"github.com/noueii/no-frame-works/internal/modules/user"
)

// Execute retrieves a post by ID and resolves the author name.
func Execute(
	ctx context.Context,
	repo post.PostRepository,
	userAPI user.UserAPI,
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

	author, err := userAPI.GetUser(ctx, user.GetUserRequest{ID: found.AuthorID})
	if err != nil {
		return post.PostView{}, fmt.Errorf("failed to resolve author: %w", err)
	}

	return post.PostView{
		ID:         found.ID,
		Title:      found.Title,
		Content:    found.Content,
		AuthorID:   found.AuthorID,
		AuthorName: author.Name,
	}, nil
}
