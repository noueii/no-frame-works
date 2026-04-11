package service

import (
	"context"

	"github.com/go-errors/errors"
	"github.com/noueii/no-frame-works/internal/core/actor"
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
) (*post.PostView, error) {
	return createpost.CreatePost(ctx, s.repo, req)
}

func (s *Service) GetPost(ctx context.Context, req post.GetPostRequest) (*post.PostView, error) {
	return getpost.GetPost(ctx, s.repo, req)
}

func (s *Service) ListAllPosts(ctx context.Context) ([]post.PostView, error) {
	a := actor.ActorFrom(ctx)
	if a == nil {
		return nil, post.ErrUnauthorized
	}

	posts, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to list all posts: %w", err)
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

func (s *Service) UpdatePost(ctx context.Context, req post.UpdatePostRequest) (*post.PostView, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, errors.Errorf("failed to find post: %w", err)
	}
	if existing == nil {
		return nil, post.ErrPostNotFound
	}

	if err := req.CheckPermission(ctx, existing); err != nil {
		return nil, err
	}

	existing.Title = req.Title
	existing.Content = req.Content

	updated, err := s.repo.Update(ctx, *existing)
	if err != nil {
		return nil, errors.Errorf("failed to update post: %w", err)
	}

	return &post.PostView{
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
		return errors.Errorf("failed to find post: %w", err)
	}
	if existing == nil {
		return post.ErrPostNotFound
	}

	if err := req.CheckPermission(ctx, existing); err != nil {
		return err
	}

	return s.repo.Delete(ctx, req.ID)
}

func (s *Service) ListPosts(
	ctx context.Context,
	req post.ListPostsRequest,
) ([]post.PostView, error) {
	return listposts.ListPosts(ctx, s.repo, req)
}
