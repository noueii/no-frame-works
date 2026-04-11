package post

import (
	"context"
	"database/sql"

	"github.com/go-errors/errors"
	jet "github.com/go-jet/jet/v2/postgres"

	"github.com/noueii/no-frame-works/db/no_frame_works/public/model"
	"github.com/noueii/no-frame-works/db/no_frame_works/public/table"
	"github.com/noueii/no-frame-works/internal/modules/post/domain"
)

func (r *PostgresPostRepository) FindByID(ctx context.Context, id string) (*domain.Post, error) {
	stmt := jet.SELECT(table.Post.AllColumns).
		FROM(table.Post).
		WHERE(table.Post.ID.EQ(jet.String(id)))

	var dest model.Post
	err := stmt.QueryContext(ctx, r.db, &dest)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrPostNotFound
	}
	if err != nil {
		return nil, errors.Errorf("query post by id: %w", err)
	}

	return toDomain(dest), nil
}
