package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, name, email, created_at, updated_at FROM "user" WHERE id = $1`,
		id,
	)

	var u domain.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query user by id: %w", err)
	}

	return &u, nil
}
