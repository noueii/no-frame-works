package editusername

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/user"
	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

// Execute changes a user's username.
func Execute(
	ctx context.Context,
	repo user.UserRepository,
	req user.EditUsernameRequest,
) (user.UserView, error) {
	if err := req.Validate(); err != nil {
		return user.UserView{}, err
	}

	existing, err := repo.FindByID(ctx, req.UserID)
	if err != nil {
		return user.UserView{}, fmt.Errorf("failed to find user: %w", err)
	}
	if existing == nil {
		return user.UserView{}, domain.ErrUserNotFound
	}

	taken, err := repo.FindByUsername(ctx, req.Username)
	if err != nil {
		return user.UserView{}, fmt.Errorf("failed to check username: %w", err)
	}
	if taken != nil && taken.ID != req.UserID {
		return user.UserView{}, domain.ErrUsernameTaken
	}

	existing.Username = req.Username
	updated, err := repo.Update(ctx, *existing)
	if err != nil {
		return user.UserView{}, fmt.Errorf("failed to update user: %w", err)
	}

	return user.UserView{
		ID:       updated.ID,
		Username: updated.Username,
		Email:    updated.Email,
	}, nil
}
