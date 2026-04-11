package user

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

func (r *Repository) Update(ctx context.Context, u domain.User) (*domain.User, error) {
	traits := map[string]interface{}{
		"email":    u.Email,
		"username": u.Username,
	}

	detail, err := r.identity.UpdateTraits(ctx, u.ID, traits)
	if err != nil {
		return nil, fmt.Errorf("update identity traits: %w", err)
	}

	return toDomain(detail), nil
}
