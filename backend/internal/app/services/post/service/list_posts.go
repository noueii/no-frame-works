package service

import (
	"context"

	"github.com/noueii/no-frame-works/internal/app/domain"
	"github.com/noueii/no-frame-works/internal/app/services/post"
)

// ListPosts delegates to ListPostsRequest.Run.
func (s *Service) ListPosts(
	ctx context.Context,
	req post.ListPostsRequest,
) ([]domain.Post, error) {
	return req.Run(ctx, s.repo)
}
