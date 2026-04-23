package service

import (
	"context"

	"github.com/noueii/no-frame-works/internal/app/domain"
	"github.com/noueii/no-frame-works/internal/app/services/post"
)

// UpdatePost delegates to UpdatePostRequest.Run, which fetches the existing
// post, mutates the requested fields, and saves the complete model back.
func (s *Service) UpdatePost(
	ctx context.Context,
	req post.UpdatePostRequest,
) (*domain.Post, error) {
	return req.Run(ctx, s.repo)
}
