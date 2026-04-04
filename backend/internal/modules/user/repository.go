package user

import (
	"context"

	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

// UserRepository defines the data access contract for the user module.
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
	Create(ctx context.Context, user domain.User) (*domain.User, error)
}
