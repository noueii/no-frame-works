# Transactions & Inter-Service Calls

How services share a transaction when one service method orchestrates multiple writes — possibly across other services.

## Primitive

Lives at `internal/app/core/tx/`. Single manager, threaded via `context.Context`.

```go
// internal/app/core/tx/manager.go
type Manager struct { db *sql.DB }

type txKey struct{}

func (m *Manager) Do(ctx context.Context, fn func(context.Context) error) error {
    if _, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
        return fn(ctx) // nested: reuse the parent tx (no-op)
    }
    tx, err := m.db.BeginTx(ctx, nil)
    if err != nil { return errors.Errorf("tx.Do: begin: %w", err) }

    ctx = context.WithValue(ctx, txKey{}, tx)
    if err := fn(ctx); err != nil {
        _ = tx.Rollback()
        return err
    }
    return tx.Commit()
}

// DB returns the active *sql.Tx if one is in context, else the pool.
func DB(ctx context.Context, def qrm.DB) qrm.DB {
    if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok { return tx }
    return def
}
```

## Exposure

The manager is an app-level infra primitive, not a service.

- **Accessed as `s.app.Tx()`** — a method on `*config.App` that returns `*tx.Manager`.
- **NOT exposed via `app.API()`** — the API registry is for services only.
- **Service shape is unchanged:** still `type Service struct { app *config.App; repo XRepository }`. No new field.

## Rules

### 1. Tx opens at the service layer — never the handler

Handlers stay thin. They do not call `s.app.Tx().Do(...)`. When atomicity is required, the service method owns the `Do` call.

**Wrong:**
```go
// handler
h.app.Tx().Do(r.Context(), func(ctx context.Context) error {
    return h.app.API().Post.Create(ctx, req)
})
```

**Right:**
```go
// handler
return h.app.API().Post.Create(r.Context(), req)

// service
func (s *Service) CreatePost(ctx context.Context, req CreateRequest) (*PostView, error) {
    var out *PostView
    err := s.app.Tx().Do(ctx, func(ctx context.Context) error {
        if err := s.repo.Insert(ctx, post); err != nil { return err }
        return s.app.API().Stats.IncrementPostCount(ctx, post.AuthorID)
    })
    return out, err
}
```

### 2. Nested `Do` is a no-op

If service A opens a tx and calls service B (which also calls `s.app.Tx().Do(...)` in its own method), the inner `Do` reuses the parent tx and returns without committing or beginning. Partial rollback is NOT supported. If you need it, open a separate code path — do not reach for SAVEPOINT.

Consequence: services compose freely. Any service method can be called either standalone (opens its own tx) or as part of a larger atomic operation (joins the parent tx). The caller doesn't need to know.

### 3. All repo methods use `tx.DB` — reads included

Every repository function — read or write — selects its `qrm.DB` via:

```go
func (r *Repo) FindByID(ctx context.Context, id string) (*domain.Post, error) {
    db := tx.DB(ctx, r.db)
    // go-jet query .Query(db, ...) / .Exec(db, ...)
}
```

This gives **read-your-writes inside a tx**, which is almost always what you want. A read inside `tx.Do(...)` MUST see writes earlier in the same `Do`.

### 4. Cross-service atomic work goes through `app.API()`, not repos

The "no direct cross-service repo access" rule still holds. Atomicity is provided by the shared context, not by exposing another service's repo.

```go
// inside service.CreatePost, already inside tx.Do
s.app.API().Stats.IncrementPostCount(ctx, authorID) // joins the same tx via ctx
```

Forbidden:
```go
s.statsRepo.Increment(ctx, authorID) // compile-impossible: App has no Repos() accessor
```

### 5. `Tx().Do` only when needed

If a service method does exactly one repo write and no cross-service calls, don't wrap it in `Do` — the pool handles the single-statement case. Reserve `Do` for:
- multiple writes in the same service
- a write plus a cross-service call that also writes
- any operation where partial success is unsafe

### 6. Context propagation is mandatory

Every service and repo method takes `ctx context.Context` as its first argument. Losing the context loses the tx. No background contexts, no `context.TODO()` inside a `Do` callback.

## Smell Checklist — stop if you're about to…

- Open a tx in a handler.
- Take `*sql.Tx` or `qrm.DB` as a method argument (tx lives in context, only).
- Reach for SAVEPOINT / nested partial rollback.
- Skip `tx.DB(ctx, r.db)` in a "read-only" repo method.
- Call another service's repo directly to "keep the tx simple."
- Pass `context.Background()` into a repo from inside a `Do` callback.
- Add `txMgr` as a field on `Service` — use `s.app.Tx()`.
