package post

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/post/domain"
)

func (r *PostgresPostRepository) FindByID(ctx context.Context, id string) (*domain.Post, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, title, content, author_id, created_at, updated_at FROM "post" WHERE id = $1`,
		id,
	)

	var p domain.Post
	err := row.Scan(&p.ID, &p.Title, &p.Content, &p.AuthorID, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query post by id: %w", err)
	}

	return &p, nil
}
