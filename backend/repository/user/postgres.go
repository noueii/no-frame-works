package user

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/app/apperrors"
	"github.com/noueii/no-frame-works/internal/app/domain"
	usermod "github.com/noueii/no-frame-works/internal/app/services/user"
)

// PostgresUserRepository is a stub implementation of usermod.UserRepository.
//
// The oapi contract for this project is scoped to posts — there is no users
// table in the schema yet — so this repo keeps in-memory state instead of
// hitting the database. It is goroutine-safe so cross-module calls from the
// HTTP handler's request goroutines don't race on the counter map.
//
// The stub returns a user for any non-empty ID. FindByID pulls the post count
// from the in-memory map; IncrementPostCount adds to it. This is enough to
// prove the god-App wiring works end-to-end (you can observe post creation
// mutating user state through the stub's internal map), without requiring a
// migration or a real users table.
type PostgresUserRepository struct {
	db *sql.DB

	mu         sync.Mutex
	postCounts map[string]int
}

// New creates a new PostgresUserRepository. The db parameter is kept for
// signature parity with PostgresPostRepository; the stub does not use it.
func New(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db:         db,
		postCounts: make(map[string]int),
	}
}

// Compile-time check that *PostgresUserRepository satisfies usermod.UserRepository.
var _ usermod.UserRepository = (*PostgresUserRepository)(nil)

func (r *PostgresUserRepository) FindByID(_ context.Context, id string) (*domain.User, error) {
	if id == "" {
		return nil, nil
	}
	r.mu.Lock()
	count := r.postCounts[id]
	r.mu.Unlock()

	now := time.Now().UTC()
	return &domain.User{
		ID:            id,
		Email:         "stub@example.com",
		NumberOfPosts: count,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (r *PostgresUserRepository) IncrementPostCount(_ context.Context, userID string) error {
	if userID == "" {
		return errors.Errorf("repo.user.IncrementPostCount: empty userID: %w", apperrors.ErrValidation)
	}
	r.mu.Lock()
	r.postCounts[userID]++
	r.mu.Unlock()
	return nil
}
