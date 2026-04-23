# Sentinel Errors

This doc explains sentinel errors from first principles — what they are, why we use them, and how they flow through the layers. If you've been writing `errors.Errorf("post not found")` inline and wondering if there's a better way, this is the better way.

## The Problem

A service function can fail in many ways. Some failures are **categorical** — "the post doesn't exist," "the actor isn't allowed to do this" — and the caller needs to do different things depending on which one happened. The handler needs to return **404** for one, **403** for another, **500** for a random database hiccup.

How does the handler tell them apart?

### The naive way (don't do this)

```go
// service
return nil, errors.Errorf("post not found")

// handler
if err.Error() == "post not found" {
    return oapi.GetPost404JSONResponse{...}, nil
}
```

String matching is fragile. Change the message, lose the match. Wrap the error with more context, lose the match. A typo breaks the handler silently.

### The right way: sentinel errors

A **sentinel error** is a pre-declared, named error **value** that you compare against by identity rather than by string content:

```go
// declared once
var ErrPostNotFound = errors.Errorf("post not found")

// returned from the service
return nil, ErrPostNotFound

// checked in the handler
if errors.Is(err, ErrPostNotFound) {
    return oapi.GetPost404JSONResponse{...}, nil
}
```

Now the handler matches on the error's **identity**, not its message. Change the message, rename the variable, wrap it — `errors.Is` still returns `true`. This is what makes sentinel errors a reliable cross-layer communication channel.

## Where Sentinels Live

### Declared in the domain

Each module's domain-level failure modes live in `backend/internal/modules/<module>/domain/errors.go`. From the `post` module:

```go
package domain

import "github.com/go-errors/errors"

var (
    ErrPostNotFound = errors.Errorf("post not found")
    ErrUnauthorized = errors.Errorf("unauthorized: no actor in context")
    ErrForbidden    = errors.Errorf("forbidden: insufficient permissions")
)
```

They live in `domain` because they describe **domain concepts**. "Post not found" is a fact about the post module's business rules, not about SQL or HTTP.

Note: we use `github.com/go-errors/errors`, not the stdlib. The `errors.Errorf` function produces an error with a stack trace attached — useful for logging. Never use `fmt.Errorf` or `errors.New` in this project.

### Re-exported at the module level

Handlers need to match these errors with `errors.Is`, but handlers **must not import the domain package** (see the type isolation rules in [00-overview.md](00-overview.md)). So the module root re-exports them:

```go
// backend/internal/modules/post/errors.go
package post

import (
    "github.com/go-errors/errors"

    "github.com/noueii/no-frame-works/internal/modules/post/domain"
)

// Domain errors re-exported for use by handler layer.
var (
    ErrPostNotFound = domain.ErrPostNotFound
    ErrUnauthorized = domain.ErrUnauthorized
    ErrForbidden    = domain.ErrForbidden
)

var (
    ErrTitleRequired    = errors.Errorf("title is required")
    ErrContentRequired  = errors.Errorf("content is required")
    ErrAuthorIDRequired = errors.Errorf("author_id is required")
    ErrIDRequired       = errors.Errorf("id is required")
)
```

Two kinds of errors appear here:

1. **Re-exports from `domain`** (`ErrPostNotFound`, etc.). These are `var X = domain.X` — the variable in `post` is **the same value** as the one in `domain`. `errors.Is(err, post.ErrPostNotFound)` and `errors.Is(err, domain.ErrPostNotFound)` both return true for the same error.
2. **Validation errors** (`ErrTitleRequired`, etc.) declared fresh in the module. These don't live in domain because they describe *request shape* rather than domain rules — they're raised by `req.Validate()`, which is defined in `api.go`.

Handlers import `post`, not `post/domain`. The type isolation stays intact.

## How Errors Flow Through the Layers

Here is the full path of an `ErrPostNotFound` from the database all the way up to an HTTP 404 response.

### 1. Repository produces it

When a query yields zero rows for a "find by ID" operation, the repository returns the domain sentinel directly:

```go
// backend/repository/post/find_by_id.go (shape)
func (r *PostgresPostRepository) FindByID(ctx context.Context, id string) (*domain.Post, error) {
    // ... run query ...
    if err == qrm.ErrNoRows {
        return nil, domain.ErrPostNotFound
    }
    if err != nil {
        return nil, errors.Errorf("find post by id: %w", err)
    }
    return toDomain(dest), nil
}
```

**Crucial rule:** `ErrPostNotFound` is returned **unwrapped**. It is the value itself, not a wrapper. This is not accidental — see the AGENTS.md rule *"Return sentinel errors directly — don't wrap domain errors"*.

### 2. Service passes it through

When the repository returns a sentinel, the service returns it **as-is**:

```go
// service
existing, err := s.repo.FindByID(ctx, req.ID)
if err != nil {
    return nil, err   // ← not wrapped
}
```

If the service wrapped this with `errors.Errorf("update failed: %w", err)`, `errors.Is` would still work (because `%w` preserves the chain), but the convention is to preserve the direct identity. Wrapping is reserved for **infrastructure errors** — random database failures, context cancellations — where the wrap adds useful debugging context:

```go
updated, err := s.repo.Update(ctx, *existing)
if err != nil {
    return nil, errors.Errorf("failed to update post: %w", err)  // ← wrap non-sentinel infra errors
}
```

The rule in one sentence: **sentinels go direct; everything else gets wrapped**.

### 3. Handler maps it to HTTP

```go
if errors.Is(err, post.ErrPostNotFound) {
    return oapi.GetPost404JSONResponse{Error: "post not found"}, nil
}
if errors.Is(err, post.ErrForbidden) {
    return oapi.GetPost403JSONResponse{Error: "forbidden"}, nil
}
if errors.Is(err, post.ErrUnauthorized) {
    return oapi.GetPost401JSONResponse{Error: "unauthorized"}, nil
}
return oapi.GetPost500JSONResponse{Error: "internal error"}, nil
```

The handler is the **only** place that translates error identity into HTTP status codes. Every sentinel gets its own branch; everything else falls through to 500.

## `errors.Is` Unpacked

`errors.Is(err, target)` returns true when `err` is equal to `target` **or** when `err` wraps a chain that contains `target`. The chain is walked via the `Unwrap()` method the error chain exposes.

```go
direct := post.ErrPostNotFound
errors.Is(direct, post.ErrPostNotFound)         // true

wrapped := errors.Errorf("load: %w", post.ErrPostNotFound)
errors.Is(wrapped, post.ErrPostNotFound)        // true — %w preserves identity

stringified := errors.Errorf("load: %s", post.ErrPostNotFound)
errors.Is(stringified, post.ErrPostNotFound)    // false — %s loses identity
```

This is why we always use `%w`, never `%v` or `%s`, when formatting an error into another error:

```go
// YES
return nil, errors.Errorf("failed to update post: %w", err)

// NO — breaks errors.Is
return nil, errors.Errorf("failed to update post: %v", err)
```

## When to Create a New Sentinel

Create a new sentinel when:

- The handler needs to make a **decision** based on it (typically: return a specific HTTP status).
- It represents a **domain-level** concept rather than an infrastructure failure.
- The same meaning is produced from multiple call sites — if every call site builds its own `errors.Errorf("not found")`, they can't be matched as one category.

Don't create a sentinel for a one-off "connection timed out while calling Stripe" error. That's an infrastructure failure — wrap it with context and let it fall through to the 500 branch.

## Common Mistakes

- **Wrapping a sentinel with `%w` when returning it.** It still works with `errors.Is`, but the convention is to preserve direct identity so grep for `return ErrX` finds every call site.
- **Using `%s` or `%v` to embed an error.** Breaks the error chain. Always `%w`.
- **Creating a new `errors.Errorf("not found")` in each call site.** Use the shared sentinel. Two independently-created errors with the same message are **not equal** under `errors.Is`.
- **Importing `"errors"` from stdlib.** Use `github.com/go-errors/errors` — it has `Is` and `Errorf` plus stack traces.
- **Declaring module-facing sentinels directly in `post/errors.go` without a domain counterpart.** Domain concepts belong in `domain/errors.go`; `post/errors.go` is a re-export layer plus request-validation errors. If an error is about the business rules of "a post," it belongs in domain.

## Mental Model

Sentinel errors are **the module's vocabulary for failure**. When you're designing a new operation, the question is: "what categories of failure does the outside world need to distinguish?" Each answer is one sentinel, declared once, referenced by identity everywhere it appears. Everything else — the random SQL errors, the context cancellations, the parse failures — is "unexpected" and gets wrapped with context and logged, not matched on.

If you're writing a service function and you're tempted to type `errors.Errorf("something"),` — stop and ask: should this be a sentinel? If the handler will ever need to distinguish this failure, the answer is yes.
