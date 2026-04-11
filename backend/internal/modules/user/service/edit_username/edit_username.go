package editusername

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/modules/user"
	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

// EditUsername changes a user's username.
func EditUsername(
	ctx context.Context,
	repo user.Repository,
	req user.EditUsernameRequest,
) (*user.View, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if err := req.CheckPermission(ctx); err != nil {
		return nil, err
	}

	existing, err := repo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.Errorf("failed to find user: %w", err)
	}
	if existing == nil {
		return nil, domain.ErrUserNotFound
	}

	taken, err := repo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.Errorf("failed to check username: %w", err)
	}
	if taken != nil && taken.ID != req.UserID {
		return nil, domain.ErrUsernameTaken
	}

	existing.Username = req.Username
	updated, err := repo.Update(ctx, *existing)
	if err != nil {
		return nil, errors.Errorf("failed to update user: %w", err)
	}

	return &user.View{
		ID:       updated.ID,
		Username: updated.Username,
		Email:    updated.Email,
	}, nil
}
