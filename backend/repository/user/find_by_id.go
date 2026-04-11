package user

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

func (r *Repository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	detail, err := r.identity.GetIdentity(ctx, id)
	if err != nil {
		return nil, errors.Errorf("get identity: %w", err)
	}

	if detail == nil {
		return nil, domain.ErrUserNotFound
	}

	return toDomain(detail), nil
}
