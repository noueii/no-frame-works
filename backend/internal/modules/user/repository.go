package user

import (
	"context"

	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

// UserRepository defines the data access contract for the user module.
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	UpdateUsername(ctx context.Context, id string, username string) (*domain.User, error)
}
