package listposts

import (
	"context"

	"github.com/go-errors/errors"
	"github.com/noueii/no-frame-works/internal/modules/post"
)

// ListPosts lists all posts by a given author.
func ListPosts(
	ctx context.Context,
	repo post.PostRepository,
	req post.ListPostsRequest,
) ([]post.PostView, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if err := req.CheckPermission(ctx); err != nil {
		return nil, err
	}

	posts, err := repo.ListByAuthor(ctx, req.AuthorID)
	if err != nil {
		return nil, errors.Errorf("failed to list posts: %w", err)
	}

	views := make([]post.PostView, len(posts))
	for i, p := range posts {
		views[i] = post.PostView{
			ID:       p.ID,
			Title:    p.Title,
			Content:  p.Content,
			AuthorID: p.AuthorID,
		}
	}

	return views, nil
}
