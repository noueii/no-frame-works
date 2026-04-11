package post

import (
	"context"

	"github.com/go-errors/errors"
	jet "github.com/go-jet/jet/v2/postgres"

	"github.com/noueii/no-frame-works/db/no_frame_works/public/model"
	"github.com/noueii/no-frame-works/db/no_frame_works/public/table"
	"github.com/noueii/no-frame-works/internal/modules/post/domain"
)

func (r *PostgresPostRepository) Update(ctx context.Context, p domain.Post) (*domain.Post, error) {
	update := toModel(p)

	stmt := table.Post.UPDATE(
		table.Post.Title,
		table.Post.Content,
		table.Post.UpdatedAt,
	).MODEL(update).
		SET(table.Post.UpdatedAt.SET(jet.TimestampExp(jet.RawTimestamp("now()")))).
		WHERE(table.Post.ID.EQ(jet.String(p.ID))).
		RETURNING(table.Post.AllColumns)

	var dest model.Post
	err := stmt.QueryContext(ctx, r.db, &dest)
	if err != nil {
		return nil, errors.Errorf("update post: %w", err)
	}

	return toDomain(dest), nil
}
