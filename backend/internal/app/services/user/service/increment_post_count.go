package service

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/services/user"
)

// IncrementPostCount is the cross-service entry point called by post.Service
// after a successful post creation. Keeping it as a first-class method on
// the user service (rather than exposing a raw repo update) means any
// future invariants — caps, audit logging, notification triggers, event
// emission — get enforced in one place.
//
// Takes *user.IncrementPostCountOp directly. The op is the input the
// caller built; there's no input→op wrapping in here.
func (s *Service) IncrementPostCount(ctx context.Context, op *user.IncrementPostCountOp) error {
	if err := op.Validate(); err != nil {
		return errors.Errorf("service.user.IncrementPostCount: validate: %w", err)
	}
	if err := s.repo.IncrementPostCount(ctx, op.UserID); err != nil {
		return errors.Errorf("service.user.IncrementPostCount: repo increment id=%s: %w", op.UserID, err)
	}
	return nil
}
