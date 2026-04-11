package user

import (
	"context"

	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

type Repository interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, u domain.User) (*domain.User, error)
}
