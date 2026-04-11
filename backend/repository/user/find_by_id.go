package user

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

func (r *Repository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	detail, err := r.identity.GetIdentity(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get identity: %w", err)
	}

	return toDomain(detail), nil
}
