package post

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/post/domain"
)

func (r *PostgresPostRepository) Create(ctx context.Context, p domain.Post) (*domain.Post, error) {
	row := r.db.QueryRowContext(
		ctx,
		`INSERT INTO "post" (title, content, author_id) VALUES ($1, $2, $3) RETURNING id, title, content, author_id, created_at, updated_at`,
		p.Title,
		p.Content,
		p.AuthorID,
	)

	var created domain.Post
	err := row.Scan(
		&created.ID,
		&created.Title,
		&created.Content,
		&created.AuthorID,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert post: %w", err)
	}

	return &created, nil
}
