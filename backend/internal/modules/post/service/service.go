package service

import (
	"context"

	"github.com/noueii/no-frame-works/internal/modules/post"
	createpost "github.com/noueii/no-frame-works/internal/modules/post/service/create_post"
	getpost "github.com/noueii/no-frame-works/internal/modules/post/service/get_post"
	listposts "github.com/noueii/no-frame-works/internal/modules/post/service/list_posts"
)

// Service implements post.PostAPI.
type Service struct {
	repo post.PostRepository
}

// New creates a new post service.
func New(repo post.PostRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreatePost(
	ctx context.Context,
	req post.CreatePostRequest,
) (post.PostView, error) {
	return createpost.Execute(ctx, s.repo, req)
}

func (s *Service) GetPost(ctx context.Context, req post.GetPostRequest) (post.PostView, error) {
	return getpost.Execute(ctx, s.repo, req)
}

func (s *Service) ListPosts(
	ctx context.Context,
	req post.ListPostsRequest,
) ([]post.PostView, error) {
	return listposts.Execute(ctx, s.repo, req)
}
