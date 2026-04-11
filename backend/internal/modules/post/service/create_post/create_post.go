package createpost

import (
	"context"

	"github.com/go-errors/errors"
	"github.com/noueii/no-frame-works/internal/modules/post"
	"github.com/noueii/no-frame-works/internal/modules/post/domain"
)

// CreatePost creates a new post.
func CreatePost(
	ctx context.Context,
	repo post.PostRepository,
	req post.CreatePostRequest,
) (*post.PostView, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if err := req.CheckPermission(ctx); err != nil {
		return nil, err
	}

	newPost := domain.Post{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: req.AuthorID,
	}

	created, err := repo.Create(ctx, newPost)
	if err != nil {
		return nil, errors.Errorf("failed to create post: %w", err)
	}

	return &post.PostView{
		ID:       created.ID,
		Title:    created.Title,
		Content:  created.Content,
		AuthorID: created.AuthorID,
	}, nil
}
