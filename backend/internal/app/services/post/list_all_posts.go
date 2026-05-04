package post

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/domain"
)

func (s *Service) ListAllPosts(ctx context.Context) ([]domain.Post, error) {
	posts, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, errors.Errorf("service.post.ListAllPosts: repo list: %w", err)
	}
	return posts, nil
}
