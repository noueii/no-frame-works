package user

import (
	"database/sql"

	"github.com/noueii/no-frame-works/internal/modules/user"
)

// Postgres implements user.UserRepository using PostgreSQL.
type Postgres struct {
	db *sql.DB
}

// New creates a new Postgres user repository.
func New(db *sql.DB) user.UserRepository {
	return &Postgres{db: db}
}
