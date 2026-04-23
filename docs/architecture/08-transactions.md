# Transactions

When a single business operation has to write to **multiple repositories atomically** — create a post and bump a counter, transfer between two accounts, accept an invitation and create a membership — the writes must succeed or fail together. The `TxManager` is how we do that without leaking transaction plumbing into every layer.

The pattern is ported from the `ck-tadasi` project's `internal/app/shared` package. It's simple, battle-tested, and doesn't require a DI framework or an ORM transaction abstraction.

## The Shape

Three pieces work together:

1. **`TxManager`** — a service-layer collaborator that wraps `*sql.DB` and knows how to start, commit, and roll back transactions.
2. **Transaction in context** — when a transaction is active, it's stashed on `context.Context` under a private key.
3. **`GetExecutor(ctx, db)`** — a repository-layer helper that returns the right executor (`*sql.Tx` if one is active, otherwise the raw `*sql.DB`). Both types satisfy `qrm.DB`, which is what go-jet needs.

The effect: **repositories stay transaction-agnostic**. They never know or care whether they're running inside a transaction.

## The TxManager

Lives in `backend/internal/core/shared/tx.go` (to be added from `ck-tadasi`'s `internal/app/shared/tx.go`):

```go
package shared

import (
    "context"
    "database/sql"

    "github.com/go-jet/jet/v2/qrm"
)

type txContextKey struct{}

type TxManager struct {
    db *sql.DB
}

func NewTxManager(db *sql.DB) *TxManager {
    return &TxManager{db: db}
}

// WithTransaction executes fn within a new transaction.
// Commits on success, rolls back on any error returned from fn.
func (m *TxManager) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
    tx, err := m.db.BeginTx(ctx, nil)
    if err != nil {
        return errors.Errorf("begin tx: %w", err)
    }

    ctx = context.WithValue(ctx, txContextKey{}, tx)

    if fnErr := fn(ctx); fnErr != nil {
        _ = tx.Rollback()
        return fnErr
    }

    return tx.Commit()
}

// EnsureTransaction runs fn within a transaction.
// If a transaction already exists in ctx, fn reuses it (no nested tx).
// If not, a new one is started.
func (m *TxManager) EnsureTransaction(ctx context.Context, fn func(context.Context) error) error {
    if HasTransaction(ctx) {
        return fn(ctx)
    }
    return m.WithTransaction(ctx, fn)
}

// GetExecutor returns the tx from context if present, otherwise the db.
// Both *sql.DB and *sql.Tx satisfy qrm.DB.
func GetExecutor(ctx context.Context, db qrm.DB) qrm.DB {
    if tx, ok := ctx.Value(txContextKey{}).(*sql.Tx); ok {
        return tx
    }
    return db
}

func HasTransaction(ctx context.Context) bool {
    _, ok := ctx.Value(txContextKey{}).(*sql.Tx)
    return ok
}
```

Four functions. `WithTransaction` and `EnsureTransaction` are called by services; `GetExecutor` is called by repositories; `HasTransaction` is rarely called directly.

## Using It From a Service

A service that needs an atomic multi-repo write takes the `TxManager` as a constructor dependency:

```go
type Service struct {
    postRepo    post.Repository
    counterRepo counter.Repository
    tx          *shared.TxManager
}

func New(postRepo post.Repository, counterRepo counter.Repository, tx *shared.TxManager) *Service {
    return &Service{
        postRepo:    postRepo,
        counterRepo: counterRepo,
        tx:          tx,
    }
}
```

Then wraps the critical section in `WithTransaction`:

```go
func (s *Service) CreatePost(ctx context.Context, req post.CreatePostRequest) (*post.View, error) {
    if err := req.Validate(); err != nil {
        return nil, err
    }
    if err := req.CheckPermission(ctx); err != nil {
        return nil, err
    }

    var created *domain.Post
    err := s.tx.WithTransaction(ctx, func(ctx context.Context) error {
        newPost := domain.Post{
            Title:    req.Title,
            Content:  req.Content,
            AuthorID: req.AuthorID,
        }

        var err error
        created, err = s.postRepo.Create(ctx, newPost)
        if err != nil {
            return err
        }

        if err := s.counterRepo.IncrementForAuthor(ctx, req.AuthorID); err != nil {
            return err
        }

        return nil
    })
    if err != nil {
        return nil, errors.Errorf("failed to create post: %w", err)
    }

    return &post.View{
        ID:       created.ID,
        Title:    created.Title,
        Content:  created.Content,
        AuthorID: created.AuthorID,
    }, nil
}
```

Three things to notice:

1. **`Validate` and `CheckPermission` run outside the transaction.** You don't want to hold a DB connection while validating input. The transaction wraps only the writes.
2. **The closure captures `created` by reference.** This is the idiomatic way to "return a value from a tx closure" — declare outside, assign inside, read after.
3. **`tx.WithTransaction` commits on nil return, rolls back on any error.** No explicit commit/rollback in the closure.

## Using It From a Repository

Every repository query changes from:

```go
err := stmt.QueryContext(ctx, r.db, &dest)
```

to:

```go
err := stmt.QueryContext(ctx, shared.GetExecutor(ctx, r.db), &dest)
```

That's the whole diff. `GetExecutor` picks the transaction from context when one is active and falls back to `r.db` otherwise. The repository doesn't need to know which is which — it just runs the query.

You can (and should) make this unconditional in every repository function. Even for functions that are "never called in a transaction," using `GetExecutor` costs nothing and means you never have to remember which functions are tx-safe.

## Nested Calls: `EnsureTransaction`

Sometimes a service function that *might* be called inside a transaction needs to guarantee *it* runs in one. Example: an internal helper called from two places, one already transactional, one not.

```go
func (s *Service) publishPost(ctx context.Context, id string) error {
    return s.tx.EnsureTransaction(ctx, func(ctx context.Context) error {
        p, err := s.postRepo.FindByID(ctx, id)
        if err != nil {
            return err
        }
        p.Published = true
        _, err = s.postRepo.Update(ctx, *p)
        return err
    })
}
```

`EnsureTransaction` checks whether a tx is already on ctx:

- **Yes** → it just calls `fn(ctx)`, no new tx. The outer transaction's commit/rollback will cover this work.
- **No** → it starts a new tx, runs `fn`, commits or rolls back.

This avoids nested transactions (which Postgres doesn't support directly — you'd need savepoints) while letting helpers be called from both contexts safely.

## Rules

- **Services start transactions, repositories run inside them.** Never call `BeginTx` in a repository, never call `GetExecutor` in a service.
- **Validation and authorization run outside the transaction.** They don't touch the DB, and holding a connection while checking input is wasteful.
- **Return from the closure on error; don't commit manually.** The manager handles commit and rollback.
- **Propagate the `ctx` from the closure, not the outer one.** Inside the closure, `ctx` has the tx; outside, it doesn't. Repository calls inside the closure must use the inner `ctx`.
- **Don't return views from inside the closure's parameters.** Capture results in outer-scope variables, then build the `*View` after the closure returns cleanly. This keeps "I succeeded" and "I have a value to hand back" as separate steps.

## What Changes When This Lands

Today, repository functions in this project take `r.db` directly:

```go
err := stmt.QueryContext(ctx, r.db, &dest)
```

When the `TxManager` ships, every repository function will be updated to:

```go
err := stmt.QueryContext(ctx, shared.GetExecutor(ctx, r.db), &dest)
```

This is a mechanical change — there is no business logic affected. Existing services that don't need transactions continue to work unchanged; the executor falls back to `r.db` when no tx is in context.

Services that want atomicity gain a new dependency (`*shared.TxManager`) and wrap the critical section. Everything else stays the same.
