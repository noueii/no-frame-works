package user

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

func (r *Repository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	identities, err := r.identity.ListIdentities(ctx)
	if err != nil {
		return nil, fmt.Errorf("list identities: %w", err)
	}

	for _, detail := range identities {
		if detail.Username == username {
			return toDomain(&detail), nil
		}
	}

	return nil, nil
}
