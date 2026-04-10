package service

import (
	"context"
	"fmt"

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

func (s *Service) ListAllPosts(ctx context.Context) ([]post.PostView, error) {
	posts, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list all posts: %w", err)
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

func (s *Service) UpdatePost(ctx context.Context, req post.UpdatePostRequest) (post.PostView, error) {
	if err := req.Validate(); err != nil {
		return post.PostView{}, fmt.Errorf("validation failed: %w", err)
	}

	existing, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return post.PostView{}, fmt.Errorf("failed to find post: %w", err)
	}
	if existing == nil {
		return post.PostView{}, post.ErrPostNotFound
	}

	existing.Title = req.Title
	existing.Content = req.Content

	updated, err := s.repo.Update(ctx, *existing)
	if err != nil {
		return post.PostView{}, fmt.Errorf("failed to update post: %w", err)
	}

	return post.PostView{
		ID:       updated.ID,
		Title:    updated.Title,
		Content:  updated.Content,
		AuthorID: updated.AuthorID,
	}, nil
}

func (s *Service) DeletePost(ctx context.Context, req post.DeletePostRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	existing, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("failed to find post: %w", err)
	}
	if existing == nil {
		return post.ErrPostNotFound
	}

	return s.repo.Delete(ctx, req.ID)
}

func (s *Service) ListPosts(
	ctx context.Context,
	req post.ListPostsRequest,
) ([]post.PostView, error) {
	return listposts.Execute(ctx, s.repo, req)
}
