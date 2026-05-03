// Package tx provides the transaction manager that threads *sql.Tx through
// context.Context. Services open transactions via app.Tx().Do(...); repos
// pick up the active tx (if any) via tx.DB(ctx, r.db) at the query callsite.
//
// Nested Do is a no-op — the inner call reuses the parent tx. Partial rollback
// is not supported.
package tx

import (
	"context"
	"database/sql"

	"github.com/go-errors/errors"
	"github.com/go-jet/jet/v2/qrm"
)

// Manager owns the *sql.DB pool and opens transactions on demand.
type Manager struct {
	db *sql.DB
}

type txKey struct{}

// NewManager constructs a Manager over the given pool.
func NewManager(db *sql.DB) *Manager {
	return &Manager{db: db}
}

// Do runs fn inside a transaction. If a tx is already active in ctx, fn runs
// on the existing tx (no new BEGIN / COMMIT — no-op nesting). Otherwise a new
// tx is opened, fn is invoked with a context carrying the tx, and the tx is
// committed on nil error or rolled back on non-nil error.
func (m *Manager) Do(ctx context.Context, fn func(context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return fn(ctx)
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Errorf("core.tx.Do: begin: %w", err)
	}

	ctx = context.WithValue(ctx, txKey{}, tx)
	if err := fn(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// DB returns the active *sql.Tx from ctx if one is present, otherwise def.
// Repositories call this at every query callsite so reads and writes inside
// a Do callback land on the same tx.
func DB(ctx context.Context, def qrm.DB) qrm.DB {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return def
}
