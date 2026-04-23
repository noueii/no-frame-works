package post

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/domain"
)

// ListAllPostsRequest is the (empty) request for the unfiltered list. It
// exists so that ListAllPosts has the same shape as every other operation
// — Run on a request type — instead of being a special parameterless method.
type ListAllPostsRequest struct{}

func (r ListAllPostsRequest) Validate() error { return nil }

// Run returns every post in the repository.
func (r ListAllPostsRequest) Run(ctx context.Context, repo PostRepository) ([]domain.Post, error) {
	posts, err := repo.ListAll(ctx)
	if err != nil {
		return nil, errors.Errorf("post.ListAllPostsRequest.Run: repo list: %w", err)
	}
	return posts, nil
}
