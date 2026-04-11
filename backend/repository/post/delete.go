package post

import (
	"context"

	"github.com/go-errors/errors"
	jet "github.com/go-jet/jet/v2/postgres"

	"github.com/noueii/no-frame-works/db/no_frame_works/public/table"
)

func (r *PostgresPostRepository) Delete(ctx context.Context, id string) error {
	stmt := table.Post.DELETE().
		WHERE(table.Post.ID.EQ(jet.String(id)))

	_, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return errors.Errorf("delete post: %w", err)
	}
	return nil
}
