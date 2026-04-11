package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

func (p *Postgres) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var u domain.User
	err := p.db.QueryRowContext(ctx,
		`SELECT id, username, email, created_at, updated_at FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
