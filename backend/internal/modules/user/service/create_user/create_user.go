package createuser

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/user"
	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

// Execute creates a new user.
func Execute(
	ctx context.Context,
	repo user.UserRepository,
	req user.CreateUserRequest,
) (user.UserView, error) {
	if err := req.Validate(); err != nil {
		return user.UserView{}, fmt.Errorf("validation failed: %w", err)
	}

	newUser := domain.User{
		Name:  req.Name,
		Email: req.Email,
	}

	created, err := repo.Create(ctx, newUser)
	if err != nil {
		return user.UserView{}, fmt.Errorf("failed to create user: %w", err)
	}

	return user.UserView{
		ID:    created.ID,
		Name:  created.Name,
		Email: created.Email,
	}, nil
}
