package createpost

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/post"
	"github.com/noueii/no-frame-works/internal/modules/post/domain"
	"github.com/noueii/no-frame-works/internal/modules/user"
)

// Execute creates a new post and resolves the author name via the user module.
func Execute(
	ctx context.Context,
	repo post.PostRepository,
	userAPI user.UserAPI,
	req post.CreatePostRequest,
) (post.PostView, error) {
	if err := req.Validate(); err != nil {
		return post.PostView{}, fmt.Errorf("validation failed: %w", err)
	}

	// Verify author exists via cross-module call
	author, err := userAPI.GetUser(ctx, user.GetUserRequest{ID: req.AuthorID})
	if err != nil {
		return post.PostView{}, fmt.Errorf("failed to resolve author: %w", err)
	}

	newPost := domain.Post{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: req.AuthorID,
	}

	created, err := repo.Create(ctx, newPost)
	if err != nil {
		return post.PostView{}, fmt.Errorf("failed to create post: %w", err)
	}

	return post.PostView{
		ID:         created.ID,
		Title:      created.Title,
		Content:    created.Content,
		AuthorID:   created.AuthorID,
		AuthorName: author.Name,
	}, nil
}
