package service

import (
	"context"

	"github.com/noueii/no-frame-works/internal/app/services/post"
)

// DeletePost delegates to DeletePostRequest.Run.
func (s *Service) DeletePost(ctx context.Context, req post.DeletePostRequest) error {
	return req.Run(ctx, s.repo)
}
