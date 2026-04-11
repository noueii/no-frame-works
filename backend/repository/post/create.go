package post

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/db/no_frame_works/public/model"
	"github.com/noueii/no-frame-works/db/no_frame_works/public/table"
	"github.com/noueii/no-frame-works/internal/modules/post/domain"
)

func (r *PostgresPostRepository) Create(ctx context.Context, p domain.Post) (*domain.Post, error) {
	insert := toModel(p)

	stmt := table.Post.INSERT(
		table.Post.Title,
		table.Post.Content,
		table.Post.AuthorID,
	).MODEL(insert).
		RETURNING(table.Post.AllColumns)

	var dest model.Post
	err := stmt.QueryContext(ctx, r.db, &dest)
	if err != nil {
		return nil, errors.Errorf("insert post: %w", err)
	}

	return toDomain(dest), nil
}

func toModel(p domain.Post) model.Post {
	return model.Post{
		Title:    p.Title,
		Content:  p.Content,
		AuthorID: p.AuthorID,
	}
}

func toDomain(m model.Post) *domain.Post {
	return &domain.Post{
		ID:        m.ID.String(),
		Title:     m.Title,
		Content:   m.Content,
		AuthorID:  m.AuthorID,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
