package listposts

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/post"
	"github.com/noueii/no-frame-works/internal/modules/user"
)

// Execute lists all posts by a given author.
func Execute(
	ctx context.Context,
	repo post.PostRepository,
	userAPI user.UserAPI,
	req post.ListPostsRequest,
) ([]post.PostView, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Resolve author name once for all posts
	author, err := userAPI.GetUser(ctx, user.GetUserRequest{ID: req.AuthorID})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve author: %w", err)
	}

	posts, err := repo.ListByAuthor(ctx, req.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	views := make([]post.PostView, len(posts))
	for i, p := range posts {
		views[i] = post.PostView{
			ID:         p.ID,
			Title:      p.Title,
			Content:    p.Content,
			AuthorID:   p.AuthorID,
			AuthorName: author.Name,
		}
	}

	return views, nil
}
