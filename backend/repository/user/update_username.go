package user

import (
	"context"

	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

func (p *Postgres) UpdateUsername(ctx context.Context, id string, username string) (*domain.User, error) {
	var u domain.User
	err := p.db.QueryRowContext(ctx,
		`UPDATE users SET username = $1, updated_at = NOW() WHERE id = $2 RETURNING id, username, email, created_at, updated_at`,
		username, id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
