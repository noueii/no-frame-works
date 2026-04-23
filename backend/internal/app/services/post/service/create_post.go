package service

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/domain"
	"github.com/noueii/no-frame-works/internal/app/services/post"
	"github.com/noueii/no-frame-works/internal/app/services/user"
)

// CreatePost runs a *post.CreatePostOp end to end. The op is the input the
// caller built — there is no input→op translation here. Each step is its
// own helper on *Service taking (ctx, op), so the body reads top-to-bottom
// as a checklist; helpers carry full access to s.app and s.repo because
// they're defined in this package, which imports config.
//
// Step helpers (s.savePost, s.incrementAuthorPostCount) live below this
// function in the same file. Adding a precondition like fetchAuthor is
// one new helper plus one line in this function — and any state that
// needs to flow between steps gets a public field on post.CreatePostOp.
func (s *Service) CreatePost(ctx context.Context, op *post.CreatePostOp) (*domain.Post, error) {
	if err := op.Validate(); err != nil {
		return nil, errors.Errorf("service.post.CreatePost: validate: %w", err)
	}

	created, err := s.savePost(ctx, op)
	if err != nil {
		return nil, errors.Errorf("service.post.CreatePost: %w", err)
	}

	if err := s.incrementAuthorPostCount(ctx, op); err != nil {
		return nil, errors.Errorf("service.post.CreatePost: %w", err)
	}

	return created, nil
}

// savePost writes the post to the own repo. Pure own-service work — uses
// s.repo directly. Returns the persisted post for the orchestrator to hand
// back at the end.
func (s *Service) savePost(ctx context.Context, op *post.CreatePostOp) (*domain.Post, error) {
	created, err := s.repo.Create(ctx, domain.Post{
		Title:    op.Title,
		Content:  op.Content,
		AuthorID: op.AuthorID,
	})
	if err != nil {
		return nil, errors.Errorf("service.post.savePost: repo create: %w", err)
	}
	return created, nil
}

// incrementAuthorPostCount keeps the user's denormalized post count in sync.
// Cross-service write — must go through s.app.API().User. The user service
// has no Repos() accessor, so this is the only legal path; any future
// invariants on user (caps, audit, events) get enforced inside
// user.Service.
func (s *Service) incrementAuthorPostCount(ctx context.Context, op *post.CreatePostOp) error {
	if err := s.app.API().User.IncrementPostCount(ctx, &user.IncrementPostCountOp{
		UserID: op.AuthorID,
	}); err != nil {
		return errors.Errorf(
			"service.post.incrementAuthorPostCount: id=%s: %w",
			op.AuthorID, err,
		)
	}
	return nil
}
