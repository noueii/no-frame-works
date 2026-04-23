# The Module API Contract (`api.go`)

Every module exposes itself through a single file at its root: `backend/internal/modules/<module>/api.go`. This file is **the contract** between the module and the rest of the application. Everything outside the module — handlers, other modules, wiring code — talks to the module through the types declared here and nothing else.

If you understand `api.go`, you understand the module.

## What's In `api.go`

Four things, and nothing else:

1. The **`API` interface** — the operations the module supports.
2. The **`View` type** — the shape external callers see on success.
3. **Request structs** — one per operation, with `Validate()` and `CheckPermission()` methods.
4. Imports of `context` and `github.com/noueii/no-frame-works/internal/core/actor` for the method signatures.

Not in `api.go`: the concrete `Service` type (that's in `service/service.go`), sentinel errors (those are in `errors.go`), domain types (those are in `domain/`).

## The `API` Interface

From `backend/internal/modules/post/api.go`:

```go
type API interface {
    CreatePost(ctx context.Context, req CreatePostRequest) (*View, error)
    GetPost(ctx context.Context, req GetPostRequest) (*View, error)
    UpdatePost(ctx context.Context, req UpdatePostRequest) (*View, error)
    DeletePost(ctx context.Context, req DeletePostRequest) error
    ListAllPosts(ctx context.Context, req ListAllPostsRequest) ([]View, error)
    ListPosts(ctx context.Context, req ListPostsRequest) ([]View, error)
}
```

Every method takes `context.Context` and a single request struct. Every method returns either `*View` (single result), `[]View` (list), or `error` (no payload). No variadic arguments, no multiple-return tuples of business values, no exposed domain types.

Handlers depend on `post.API`, not `*service.Service`. This is what lets you swap the implementation in tests (with a fake) and what keeps the service's internals invisible to the rest of the app.

## The `View` Type

```go
type View struct {
    ID       string
    Title    string
    Content  string
    AuthorID string
}
```

`View` is **separate from `domain.Post`** even when the fields overlap. This separation is deliberate:

- `domain.Post` is **internal** — it has timestamps, hidden state, business rules, whatever the module needs to operate. It changes when the business rules change.
- `post.View` is **external** — it's the stable shape external callers contract against. It changes when the contract changes, which is a much less frequent and more deliberate event.

If you add a field to `domain.Post` (say, `ModeratedAt`), nothing outside the module breaks. If you add a field to `View`, every caller now sees it — that's an API change, and it should feel like one.

The service is responsible for the translation. At the end of a service function, you see:

```go
return &post.View{
    ID:       updated.ID,
    Title:    updated.Title,
    Content:  updated.Content,
    AuthorID: updated.AuthorID,
}, nil
```

Boring, explicit, and exactly the point. The repository returns a `domain.Post`; the service picks the fields the outside world should see.

## Request Structs

Each operation gets its own request struct, and that struct owns **two** methods: `Validate()` and `CheckPermission(...)`.

### `Validate()` — shape and required fields

```go
type CreatePostRequest struct {
    Title    string
    Content  string
    AuthorID string
}

func (r CreatePostRequest) Validate() error {
    if r.Title == "" {
        return ErrTitleRequired
    }
    if r.Content == "" {
        return ErrContentRequired
    }
    if r.AuthorID == "" {
        return ErrAuthorIDRequired
    }
    return nil
}
```

`Validate` is concerned with **input shape only** — required fields, lengths, formats. It does not query the database, it does not check permissions, it does not know who's calling. It can be called with zero context and return a deterministic answer.

### `CheckPermission(ctx[, model])` — authorization

There are two signatures, chosen by the operation:

**Simple (no existing model needed):**

```go
func (r CreatePostRequest) CheckPermission(ctx context.Context) error {
    a := actor.From(ctx)
    if a == nil {
        return ErrUnauthorized
    }
    return nil
}
```

**Ownership check (needs the existing model):**

```go
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

The second form is used for updates and deletes, where the permission depends on who owns the record. The service fetches the model first, then passes it in:

```go
existing, err := s.repo.FindByID(ctx, req.ID)
if err != nil {
    return nil, err
}
if permErr := req.CheckPermission(ctx, existing); permErr != nil {
    return nil, permErr
}
```

This split keeps authorization logic visible at the call site. You can read any service function top-to-bottom and see exactly what's validated, what's loaded, and what's authorized — there is no magic middleware inserting checks elsewhere.

Authorization **rules** (the boolean predicates themselves) live on domain types: `post.CanModify(a)`. The request struct's `CheckPermission` is just the dispatcher — it grabs the actor from context, asks the domain, and translates a `false` into a sentinel error. See [03-domain.md](03-domain.md).

## Why Request Structs Own Validation and Permissions

It may seem odd to put `Validate` and `CheckPermission` on the request struct rather than on the service. The reason is **co-location**: when you're reading or modifying an operation, everything you need to know about it — the input shape, the required fields, the permission rules — is within eye-shot in a single file. You don't have to hunt across a middleware chain, a validation package, and a guards package to understand what happens when `CreatePost` is called.

It also makes the contract self-documenting. Looking at `api.go`, you can see every request, every validation rule, every permission check. That's the whole public surface of the module on a single screen.

## Request Structs Are Not DTOs

Request structs are not DTOs copied from the oapi contract. They are the module's own language for "what does this operation need?" The handler's job is to translate between the two. If the oapi request has `title_html` and the service wants `title`, the handler does that renaming. This decoupling is what lets you change the oapi schema without touching the service, and vice versa.
