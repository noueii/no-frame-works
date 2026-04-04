package post

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/post/domain"
)

func (r *PostgresPostRepository) ListByAuthor(ctx context.Context, authorID string) ([]domain.Post, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, title, content, author_id, created_at, updated_at FROM "post" WHERE author_id = $1 ORDER BY created_at DESC`,
		authorID,
	)
	if err != nil {
		return nil, fmt.Errorf("query posts by author: %w", err)
	}
	defer rows.Close()

	var posts []domain.Post
	for rows.Next() {
		var p domain.Post
		if scanErr := rows.Scan(&p.ID, &p.Title, &p.Content, &p.AuthorID, &p.CreatedAt, &p.UpdatedAt); scanErr != nil {
			return nil, fmt.Errorf("scan post row: %w", scanErr)
		}
		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate post rows: %w", err)
	}

	return posts, nil
}
