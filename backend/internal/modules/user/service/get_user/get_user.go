package getuser

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/user"
)

// Execute retrieves a user by ID.
func Execute(
	ctx context.Context,
	repo user.UserRepository,
	req user.GetUserRequest,
) (user.UserView, error) {
	if err := req.Validate(); err != nil {
		return user.UserView{}, fmt.Errorf("validation failed: %w", err)
	}

	found, err := repo.FindByID(ctx, req.ID)
	if err != nil {
		return user.UserView{}, fmt.Errorf("failed to get user: %w", err)
	}

	if found == nil {
		return user.UserView{}, user.ErrUserNotFound
	}

	return user.UserView{
		ID:    found.ID,
		Name:  found.Name,
		Email: found.Email,
	}, nil
}
