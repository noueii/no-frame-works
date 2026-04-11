package service

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/modules/post"
	createpost "github.com/noueii/no-frame-works/internal/modules/post/service/create_post"
	getpost "github.com/noueii/no-frame-works/internal/modules/post/service/get_post"
	listposts "github.com/noueii/no-frame-works/internal/modules/post/service/list_posts"
)

// Service implements post.API.
type Service struct {
	repo post.Repository
}

// New creates a new post service.
func New(repo post.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreatePost(
	ctx context.Context,
	req post.CreatePostRequest,
) (*post.View, error) {
	return createpost.CreatePost(ctx, s.repo, req)
}

func (s *Service) GetPost(ctx context.Context, req post.GetPostRequest) (*post.View, error) {
	return getpost.GetPost(ctx, s.repo, req)
}

func (s *Service) ListAllPosts(
	ctx context.Context,
	req post.ListAllPostsRequest,
) ([]post.View, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if err := req.CheckPermission(ctx); err != nil {
		return nil, err
	}

	posts, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to list all posts: %w", err)
	}

	views := make([]post.View, len(posts))
	for i, p := range posts {
		views[i] = post.View{
			ID:       p.ID,
			Title:    p.Title,
			Content:  p.Content,
			AuthorID: p.AuthorID,
		}
	}

	return views, nil
}

func (s *Service) UpdatePost(
	ctx context.Context,
	req post.UpdatePostRequest,
) (*post.View, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	if permErr := req.CheckPermission(ctx, existing); permErr != nil {
		return nil, permErr
	}

	existing.Title = req.Title
	existing.Content = req.Content

	updated, err := s.repo.Update(ctx, *existing)
	if err != nil {
		return nil, errors.Errorf("failed to update post: %w", err)
	}

	return &post.View{
		ID:       updated.ID,
		Title:    updated.Title,
		Content:  updated.Content,
		AuthorID: updated.AuthorID,
	}, nil
}

func (s *Service) DeletePost(ctx context.Context, req post.DeletePostRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	existing, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return err
	}

	if permErr := req.CheckPermission(ctx, existing); permErr != nil {
		return permErr
	}

	return s.repo.Delete(ctx, req.ID)
}

func (s *Service) ListPosts(
	ctx context.Context,
	req post.ListPostsRequest,
) ([]post.View, error) {
	return listposts.ListPosts(ctx, s.repo, req)
}
