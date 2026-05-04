# Repository Layer

Repositories live in `backend/repository/<module>/`. They are the project's only gateway to the database. A repository takes a domain model, runs a query, and returns a domain model. That's the entire contract.

## Responsibilities

- Accept a **complete** domain model or a simple identifier (ID, filter struct).
- Convert to the go-jet-generated `model.*` type via a private `toModel` function.
- Execute a go-jet query — never raw SQL.
- Convert the result back via `toDomain` and return a `domain.*` type.

## What Repositories Must NOT Do

- No business logic. No "if the post is published, don't update X." That belongs in the service.
- No `oapi.*` imports. No API contract types. The repo doesn't know about HTTP or the module's external contract.
- No partial-update methods. No `UpdateTitle(id, title)`. The service hands you a complete model; you save it.
- No raw SQL strings. Use the go-jet query builder.
- No sentinel errors of its own invention. If the service needs to signal "not found," that sentinel is declared in the module's `domain/errors.go`, and the repository returns it directly when a query yields no rows.

## Example: Create

From `backend/repository/post/create.go`:

```go
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
```

Read it top to bottom: `domain.Post` comes in, `toModel` converts it to `model.Post` (the go-jet row type), the query runs, `toDomain` converts the result back, `*domain.Post` comes out. No business logic in the middle.

## Mapping Functions Live Here

`toModel` and `toDomain` are **private** to the repository package. They are not exported, they are not shared across modules, and they do nothing but field assignment. If a mapping needs logic beyond field copying, that logic belongs in the service, not here.

Two mapping functions per module is the rule. If you find yourself writing `toModelForUpdate` and `toModelForCreate`, your domain model is probably wrong — either it's missing fields, or your updates are not actually full-model saves (see below).

### Rehydration: the one place struct literals for domain types are allowed

Everywhere else in the codebase, new domain entities go through their factory (`domain.NewPost`, etc. — see [03-domain.md](03-domain.md)). `toDomain` is the exception.

```go
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
```

Why this is not a contradiction: `NewPost` is for creating a **new** entity — it generates a fresh UUID and stamps `CreatedAt = now()`. If the repository called `NewPost` when reading from the DB, it would overwrite the real ID and timestamps with freshly-generated ones. Rehydration needs the exact opposite: **copy the persisted state verbatim**. A struct literal is the only thing that gets this right.

So the rule has two halves, and both matter:

- **New entity →** `domain.NewPost(...)` (service layer)
- **Existing entity from DB →** struct literal in `toDomain` (repository layer)

These paths never cross. A service never calls `toDomain`, and the repository never calls `NewPost`. If you're ever tempted to do either, stop and re-read this section.

## Model-In, Model-Out

The repository interface uses domain models as input **and** output. Never take "a title and a content," always take "a Post."

```go
// YES
func (r *PostgresPostRepository) Update(ctx context.Context, p domain.Post) (*domain.Post, error)

// NO
func (r *PostgresPostRepository) UpdateTitle(ctx context.Context, id, title string) error
```

This is what makes fetch-mutate-save work in the service. See [02-service.md](02-service.md).

## Updates Use `MODEL()` with `MutableColumns`

For updates, use go-jet's `MODEL(mutableCopy).SET(table.T.MutableColumns)` pattern so every mutable column is written in one statement:

```go
stmt := table.Post.UPDATE(table.Post.MutableColumns).
    MODEL(toModel(p)).
    WHERE(table.Post.ID.EQ(pg.UUID(uuid.MustParse(p.ID)))).
    RETURNING(table.Post.AllColumns)
```

If you need to deviate — skip some mutable columns, update a single field — add an inline comment explaining **why**. The default is "update everything the domain thinks it owns," matching the fetch-mutate-save flow.

## One Function Per File

In sub-packages, each exported repository operation lives in its own file: `create.go`, `update.go`, `find_by_id.go`, `delete.go`. The interface is defined in the module (`post.Repository`), not in the repository package. The repository package holds the Postgres implementation of that interface.

## Error Handling

Wrap infrastructure errors with context so they can be read in logs:

```go
return nil, errors.Errorf("insert post: %w", err)
```

Return **domain sentinels** directly (unwrapped) when the query outcome maps to a domain concept — typically "not found":

```go
if err == qrm.ErrNoRows {
    return nil, domain.ErrPostNotFound
}
```

This is the one place the repository is allowed to import the domain package's error variables. The service and handler then match them with `errors.Is`. See [07-sentinel-errors.md](07-sentinel-errors.md).

## Transactions

Repositories run against whatever executor is in the context — either the raw `*sql.DB` or a `*sql.Tx` started by the `TxManager`. The helper `shared.GetExecutor(ctx, r.db)` returns the right one transparently. Once transactions are in use, the query line changes from:

```go
err := stmt.QueryContext(ctx, r.db, &dest)
```

to:

```go
err := stmt.QueryContext(ctx, shared.GetExecutor(ctx, r.db), &dest)
```

The repository stays agnostic — it doesn't know whether it's running inside a transaction or not. See [08-transactions.md](08-transactions.md) for the full pattern.
