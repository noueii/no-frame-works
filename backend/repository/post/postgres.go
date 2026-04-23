package post

import (
	"database/sql"

	postmod "github.com/noueii/no-frame-works/internal/app/services/post"
)

// PostgresPostRepository implements postmod.PostRepository using PostgreSQL with go-jet.
type PostgresPostRepository struct {
	db *sql.DB
}

// New creates a new PostgresPostRepository.
func New(db *sql.DB) *PostgresPostRepository {
	return &PostgresPostRepository{db: db}
}

// Compile-time check that PostgresPostRepository implements PostRepository.
var _ postmod.PostRepository = (*PostgresPostRepository)(nil)
