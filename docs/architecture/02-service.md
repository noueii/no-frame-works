# Service Layer

Services live in `backend/internal/modules/<module>/service/`. This is where business logic happens. Handlers orchestrate HTTP shapes; repositories orchestrate SQL; **services orchestrate behavior**.

## The Service Ritual

Every service function follows the same script:

1. `req.Validate()`
2. `req.CheckPermission(ctx[, existingModel])`
3. Business logic — typically **fetch → mutate → save** with the repository
4. Return `*View` on success, `nil, err` on failure

No middleware handles validation. No middleware handles permissions. Both live on the request struct as methods, and every service function calls them in this exact order. See [05-api-contract.md](05-api-contract.md) for the request struct shape.

## Full Example: UpdatePost

From `backend/internal/modules/post/service/service.go`:

```go
func (s *Service) UpdatePost(
    ctx context.Context,
    req post.UpdatePostRequest,
) (*post.View, error) {
    if err := req.Validate(); err != nil {
        return nil, err
    }

    existing, err := s.repo.FindByID(ctx, req.ID)
    if err != nil {
        return nil, err
    }

    if permErr := req.CheckPermission(ctx, existing); permErr != nil {
        return nil, permErr
    }

    existing.Title = req.Title
    existing.Content = req.Content

    updated, err := s.repo.Update(ctx, *existing)
    if err != nil {
        return nil, errors.Errorf("failed to update post: %w", err)
    }

    return &post.View{
        ID:       updated.ID,
        Title:    updated.Title,
        Content:  updated.Content,
        AuthorID: updated.AuthorID,
    }, nil
}
```

Notice the ordering: for operations that check permissions **against an existing model** (ownership, for example), we fetch first and then call `CheckPermission(ctx, existing)`. For operations that need only the actor (create, list), `CheckPermission(ctx)` runs immediately after `Validate`. The request struct's method signature tells you which is which.

## Fetch-Mutate-Save

For update operations, the domain model is the **source of truth**, not the request. The flow is always:

1. Fetch the complete existing model from the repository.
2. Mutate the fields in memory — only the ones the request is changing.
3. Save the complete model back.

```go
existing.Title = req.Title
existing.Content = req.Content
updated, err := s.repo.Update(ctx, *existing)
```

This pattern matters because it:
- Keeps the repository dumb. Repos don't know what "update a post" means; they know how to save a `Post`.
- Prevents accidental zero-ing of fields the client didn't send.
- Centralizes domain-level invariants in the service where they belong, not in partial SQL updates scattered through the repo.

If you find yourself writing `repo.UpdateTitle(id, title)`, stop — you are leaking service-layer concerns into the repository.

## Return Shape: `*View`, never empty structs

Services return `*post.View` (pointer) on success and `nil, err` on failure. **Never** return an empty `post.View{}, err`:

```go
// YES
return nil, errors.Errorf("failed to update post: %w", err)

// NO
return post.View{}, err
```

An empty struct looks like valid data to the caller and invites bugs where a zero-valued response is mistaken for a successful one. The pointer convention makes `nil` unambiguously "no result."

## Error Handling

Two rules, taken together, make the layer contract work:

1. **Return sentinel errors directly, unwrapped.** When a service propagates `ErrPostNotFound`, it returns it as-is so the handler can match it with `errors.Is`.
2. **Wrap everything else with `%w`.** When the repo fails with some SQL error, wrap it: `errors.Errorf("failed to update post: %w", err)`. This preserves the error chain for logs and debugging.

Never use `fmt.Errorf` or the stdlib `errors` package. Always `github.com/go-errors/errors`. See [07-sentinel-errors.md](07-sentinel-errors.md) for the full rationale.

## File Layout: One Operation Per File

The service package is flat. Every operation is a method on `*Service`, and every operation lives in its own file named after the method in snake_case:

```
backend/internal/modules/post/service/
    service.go          // the Service struct + constructor
    create_post.go      // (s *Service) CreatePost(...)
    get_post.go         // (s *Service) GetPost(...)
    update_post.go      // (s *Service) UpdatePost(...)
    delete_post.go      // (s *Service) DeletePost(...)
    list_posts.go       // (s *Service) ListPosts(...)
    list_all_posts.go   // (s *Service) ListAllPosts(...)
```

All files belong to `package service`. **There are no sub-packages.** Earlier iterations of this project nested operations into sub-packages like `service/create_post/create_post.go` — that was a mistake and should not be repeated. It added a package boundary for no semantic reason, forced operations to take the repo as an argument instead of a receiver, and broke jump-to-definition for anyone searching by method name.

`service.go` holds only the struct and constructor:

```go
package service

type Service struct {
    repo post.Repository
}

func New(repo post.Repository) *Service {
    return &Service{repo: repo}
}
```

Each operation file holds exactly one method:

```go
// backend/internal/modules/post/service/create_post.go
package service

func (s *Service) CreatePost(
    ctx context.Context,
    req post.CreatePostRequest,
) (*post.View, error) {
    if err := req.Validate(); err != nil {
        return nil, err
    }
    if err := req.CheckPermission(ctx); err != nil {
        return nil, err
    }

    newPost := domain.NewPost(req.Title, req.Content, req.AuthorID)

    created, err := s.repo.Create(ctx, newPost)
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

Naming rule: the file name matches the method name in snake_case. `GetPosts` → `get_posts.go`, `ListAllPosts` → `list_all_posts.go`. You should be able to guess the file from the method and vice versa without searching.

Note that `newPost` is built via `domain.NewPost(...)` rather than a struct literal. Outside the repository's `toDomain` helper, services **never** construct a `domain.Post` with a struct literal — always go through the factory. This keeps ID generation, timestamps, and any future invariants in one place. See [03-domain.md](03-domain.md) for the factory pattern and [04-repository.md](04-repository.md) for the one exception (rehydration from the DB).

## Dependencies

Services receive dependencies through a constructor:

```go
type Service struct {
    repo post.Repository
}

func New(repo post.Repository) *Service {
    return &Service{repo: repo}
}
```

Dependencies are **interfaces** defined in the module (like `post.Repository`), not concrete types. This keeps the service independent of repository implementations and makes it trivial to fake in tests.

## Service Access

Services are only accessible through the module's `API` interface. Nothing outside the module package should import the `service` sub-package directly. The wiring happens once, at startup, where `New(...)` returns a `*Service` that gets stored as an `API` interface in the handler struct. After that, the concrete type is invisible.
