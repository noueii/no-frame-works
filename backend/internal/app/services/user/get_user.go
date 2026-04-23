package user

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/apperrors"
	"github.com/noueii/no-frame-works/internal/app/domain"
)

// GetUserRequest is the request to fetch a user by ID.
type GetUserRequest struct {
	ID string
}

func (r GetUserRequest) Validate() error {
	if r.ID == "" {
		return apperrors.Validation(apperrors.CodeUserIDRequired, "user id is required", nil)
	}
	return nil
}

// Run validates and fetches a user by ID. Returns apperrors.NotFound when the
// user does not exist, so the handler can map it to 404 via errors.Is.
func (r GetUserRequest) Run(ctx context.Context, repo UserRepository) (*domain.User, error) {
	if err := r.Validate(); err != nil {
		return nil, errors.Errorf("user.GetUserRequest.Run: validate: %w", err)
	}
	found, err := repo.FindByID(ctx, r.ID)
	if err != nil {
		return nil, errors.Errorf("user.GetUserRequest.Run: repo find id=%s: %w", r.ID, err)
	}
	if found == nil {
		return nil, apperrors.NotFound(
			apperrors.CodeUserNotFound,
			"user not found",
			map[string]any{"user_id": r.ID},
		)
	}
	return found, nil
}
