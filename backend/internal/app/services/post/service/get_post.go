package service

import (
	"context"

	"github.com/noueii/no-frame-works/internal/app/domain"
	"github.com/noueii/no-frame-works/internal/app/services/post"
)

// GetPost delegates to GetPostRequest.Run, which validates and fetches the
// post (and returns apperrors.NotFound when missing). The wrapper exists so
// the service satisfies post.PostAPI.
func (s *Service) GetPost(ctx context.Context, req post.GetPostRequest) (*domain.Post, error) {
	return req.Run(ctx, s.repo)
}
