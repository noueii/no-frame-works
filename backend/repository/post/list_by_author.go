package post

import (
	"context"
	"fmt"

	jetpostgres "github.com/go-jet/jet/v2/postgres"

	"github.com/noueii/no-frame-works/db/no_frame_works/public/model"
	"github.com/noueii/no-frame-works/db/no_frame_works/public/table"
	"github.com/noueii/no-frame-works/internal/app/domain"
)

func (r *PostgresPostRepository) ListByAuthor(
	ctx context.Context,
	authorID string,
) ([]domain.Post, error) {
	stmt := jetpostgres.SELECT(table.Post.AllColumns).
		FROM(table.Post).
		WHERE(table.Post.AuthorID.EQ(jetpostgres.String(authorID))).
		ORDER_BY(table.Post.CreatedAt.DESC())

	var dest []model.Post
	err := stmt.QueryContext(ctx, r.db, &dest)
	if err != nil {
		return nil, fmt.Errorf("query posts by author: %w", err)
	}

	posts := make([]domain.Post, len(dest))
	for i, m := range dest {
		posts[i] = *toDomain(m)
	}

	return posts, nil
}
