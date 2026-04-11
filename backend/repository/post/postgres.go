package post

import (
	"database/sql"

	postmod "github.com/noueii/no-frame-works/internal/modules/post"
)

// PostgresPostRepository implements postmod.Repository using PostgreSQL with go-jet.
type PostgresPostRepository struct {
	db *sql.DB
}

// New creates a new PostgresPostRepository.
func New(db *sql.DB) *PostgresPostRepository {
	return &PostgresPostRepository{db: db}
}

// Compile-time check that PostgresPostRepository implements Repository.
var _ postmod.Repository = (*PostgresPostRepository)(nil)
