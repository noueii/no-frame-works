package service

import (
	"context"

	"github.com/noueii/no-frame-works/internal/app/domain"
	"github.com/noueii/no-frame-works/internal/app/services/post"
)

// ListAllPosts delegates to ListAllPostsRequest.Run. The empty request struct
// exists so this operation has the same Run shape as every other one.
func (s *Service) ListAllPosts(ctx context.Context, req post.ListAllPostsRequest) ([]domain.Post, error) {
	return req.Run(ctx, s.repo)
}
