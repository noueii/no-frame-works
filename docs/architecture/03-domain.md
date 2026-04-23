# Domain Layer

Domain code lives in `backend/internal/modules/<module>/domain/`. This is the most restricted layer in the project: it holds **pure business types and rules**, and nothing else. If the rest of the codebase burned down, the domain package should still compile.

## Responsibilities

The domain layer owns:

- **Types** that represent business concepts (e.g. `Post`, `User`).
- **Pure methods** on those types expressing business rules (e.g. `CanModify`, `IsPublished`).
- **Sentinel errors** that describe domain-level failure modes (`ErrPostNotFound`, `ErrForbidden`).

That's all. No transport, no persistence, no clocks, no random numbers, no external SDKs.

## Forbidden Imports

A domain file **must not** import:

- `database/sql`, `net/http`, or any transport/storage package
- `oapi` or API contract types
- `model` (the go-jet generated package)
- External SDKs (Stripe, S3, etc.)
- Other modules' domain packages (types are owned by their module)

If a domain method needs "the current time" or "a random ID," the caller (service) passes it in. The domain layer does not reach out.

## Example: `domain.Post`

From `backend/internal/modules/post/domain/models.go`:

```go
package domain

import (
    "time"

    "github.com/noueii/no-frame-works/internal/core/actor"
)

type Post struct {
    ID        string
    Title     string
    Content   string
    AuthorID  string
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (p Post) CanModify(a actor.Actor) bool {
    if a.IsSystem() {
        return true
    }
    if ua, ok := a.(actor.UserActor); ok && ua.HasRole(actor.RoleAdmin) {
        return true
    }
    return p.AuthorID == a.UserID().String()
}
```

A few things to notice:

- `time.Time` is fine — the standard library `time` package is a value type, not infrastructure. The domain embeds timestamps because "when was this created" is a business fact.
- `actor.Actor` is imported from `internal/core/actor`. The actor package is **core** (shared primitives), not infrastructure — it's a pure interface describing "who is acting." See [06-actor.md](06-actor.md).
- `CanModify` takes the actor as an **argument**, not from context. Domain methods never touch `context.Context`. If a method needs a collaborator, the caller hands it in. This keeps the domain testable without mocks and makes the business rule explicit at every call site.

## Constructing New Entities: `NewPost`

When a service needs a **new** domain entity (not one rehydrated from the database), it calls a factory function on the domain package rather than building a struct literal:

```go
func NewPost(title, content, authorID string) Post {
    now := time.Now().UTC()
    return Post{
        ID:        uuid.NewString(),
        Title:     title,
        Content:   content,
        AuthorID:  authorID,
        CreatedAt: now,
        UpdatedAt: now,
    }
}
```

This is the idiomatic Go factory pattern — the same shape as `http.NewRequest`, `sql.Open`, `log.New`, or `bytes.NewBuffer`. It's not OOP; it's just a function that returns a struct with its invariants satisfied.

### Why a factory instead of a struct literal

Three things live inside `NewPost`:

1. **ID generation.** The post has a UUID from the moment it exists, not "sometime after the DB round-trip." This matters because the service can reference `post.ID` for permission checks, multi-repo inserts, or validation before the first SQL statement runs.
2. **Timestamps.** `CreatedAt` and `UpdatedAt` are set to the same moment, centrally. No service has to remember to set them, and no service can accidentally leave them at their zero value.
3. **Invariants.** Future-proofing: if `Post` ever grows a required default (`Status: StatusDraft`, say), you add it once here instead of hunting through every creation site.

The rule: **outside the repository, never build a `domain.Post` with a struct literal.** Always go through `NewPost`. The compiler won't enforce this today (fields are exported), but it's a convention worth keeping because it gives you one place to change when the invariants evolve.

### When a domain type does *not* need a factory

Not every domain type needs a `New*` function. If a type has no generated fields, no timestamps, and no invariants — just a data bag like `Tag{Name: "go"}` — a struct literal is fine. The rule isn't "always use factories"; it's **"use a factory when construction has work to do."** For `Post`, the work is ID + two timestamps, which is enough to earn one.

### The one exception: rehydration from the database

When the repository reads a row from Postgres and turns it into a `domain.Post`, it **does not** call `NewPost`. It uses a plain struct literal inside its private `toDomain` helper, filling every field from the row. See [04-repository.md](04-repository.md) for why — and why this is not a contradiction. In short: `NewPost` is for **new** entities (generates ID, stamps `now`), while `toDomain` is for **existing** entities (copies the persisted ID and timestamps verbatim). The two paths never cross.

## Authorization Lives Here

Authorization rules — "who can do what to what" — are domain logic. They belong on domain types as pure methods that return `bool`:

```go
func (p Post) CanModify(a actor.Actor) bool { ... }
func (p Post) CanPublish(a actor.Actor) bool { ... }
```

The request struct's `CheckPermission` (in `api.go`) is the **dispatcher**: it pulls the actor from context and calls the domain method, translating a boolean into a sentinel error.

```go
// in post/api.go
func (r UpdatePostRequest) CheckPermission(ctx context.Context, post *domain.Post) error {
    a := actor.From(ctx)
    if a == nil {
        return ErrUnauthorized
    }
    if !post.CanModify(a) {
        return ErrForbidden
    }
    return nil
}
```

The split is deliberate: the domain method is a pure predicate you can test with a table-driven test; the request method knows about context and errors. Neither can be reused in the wrong place.

## Sentinel Errors in the Domain

Failure modes that describe domain concepts go in `domain/errors.go`:

```go
package domain

import "github.com/go-errors/errors"

var (
    ErrPostNotFound = errors.Errorf("post not found")
    ErrUnauthorized = errors.Errorf("unauthorized: no actor in context")
    ErrForbidden    = errors.Errorf("forbidden: insufficient permissions")
)
```

These are the **module's vocabulary for failure**. Other layers compare against them with `errors.Is`. Full treatment in [07-sentinel-errors.md](07-sentinel-errors.md).

## No Persistence, No Presentation

The domain type does not have `ToJSON`, `FromRow`, `ToDBModel`, or any other method that knows about a specific output format. Those concerns belong:

- In the repository: `toModel(p domain.Post) model.Post` and `toDomain(m model.Post) *domain.Post`.
- In the service: building `*post.View` from a `domain.Post` before returning.

If you add a `ToX` method on a domain type, you are merging layers.

## Module-Owned Types

Each module owns its domain types. There is **no shared domain package** like `domain/shared`. If two modules both need a concept of "money," they either each define it or there is a core primitive in `internal/core/` — but `internal/modules/post/domain/Post` and `internal/modules/user/domain/User` live in different packages and never import each other.

When a service needs information from another module, it calls that module's `API` interface — it does not reach into the other module's domain.
