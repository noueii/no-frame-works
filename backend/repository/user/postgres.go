package user

import (
	"database/sql"

	usermod "github.com/noueii/no-frame-works/internal/modules/user"
)

// PostgresUserRepository implements usermod.UserRepository using PostgreSQL.
type PostgresUserRepository struct {
	db *sql.DB
}

// New creates a new PostgresUserRepository.
func New(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Compile-time check that PostgresUserRepository implements UserRepository.
var _ usermod.UserRepository = (*PostgresUserRepository)(nil)
