package createpost

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/post"
	"github.com/noueii/no-frame-works/internal/modules/post/domain"
)

// Execute creates a new post.
func Execute(
	ctx context.Context,
	repo post.PostRepository,
	req post.CreatePostRequest,
) (post.PostView, error) {
	if err := req.Validate(); err != nil {
		return post.PostView{}, fmt.Errorf("validation failed: %w", err)
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
		ID:       created.ID,
		Title:    created.Title,
		Content:  created.Content,
		AuthorID: created.AuthorID,
	}, nil
}
