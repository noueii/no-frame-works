package post

import (
	"context"

	"github.com/noueii/no-frame-works/internal/app/domain"
)

// PostRepository defines the data access contract for the post module.
type PostRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Post, error)
	ListAll(ctx context.Context) ([]domain.Post, error)
	ListByAuthor(ctx context.Context, authorID string) ([]domain.Post, error)
	Create(ctx context.Context, post domain.Post) (*domain.Post, error)
	Update(ctx context.Context, post domain.Post) (*domain.Post, error)
	Delete(ctx context.Context, id string) error
}
