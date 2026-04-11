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
		return nil, err
	}

	taken, err := repo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// username is available, continue
		} else {
			return nil, err
		}
	} else if taken.ID != req.UserID {
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
