package user

import (
	"context"

	"github.com/go-errors/errors"
	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

func (r *Repository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	identities, err := r.identity.ListIdentities(ctx)
	if err != nil {
		return nil, errors.Errorf("list identities: %w", err)
	}

	for _, detail := range identities {
		if detail.Username == username {
			return toDomain(&detail), nil
		}
	}

	return nil, nil
}
