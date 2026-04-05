package user

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

func (r *PostgresUserRepository) Create(ctx context.Context, u domain.User) (*domain.User, error) {
	row := r.db.QueryRowContext(
		ctx,
		`INSERT INTO "user" (name, email) VALUES ($1, $2) RETURNING id, name, email, created_at, updated_at`,
		u.Name,
		u.Email,
	)

	var created domain.User
	err := row.Scan(
		&created.ID,
		&created.Name,
		&created.Email,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return &created, nil
}
