# Service Layer Review Rubric

You are reviewing service code in a Go backend with a modular architecture. Each module exposes a public API interface that is the only way to interact with the module.

Services contain **business logic**. They validate input, orchestrate operations, and return results. They are accessed exclusively through the module's API interface.

## Allowed types

Services may only work with:
- **API contract types** (request structs + view types from `api.go`) — for input/output
- **Domain models** (`domain.*`) — for internal business logic

Services must NOT import or use:
- oapi-codegen generated types (`oapi.*`)
- Database/go-jet models (`model.*`)

## Rules

### 1. Validate first, then check permission

Every service function must call `req.Validate()` first, then `req.CheckPermission(...)`. There is no separate permission middleware layer — permission checking lives on the request type and is called in the service function.

**Simple case** — no model needed for permission check:
```go
func CreatePost(ctx context.Context, repo post.PostRepository, req post.CreatePostRequest) (*post.PostView, error) {
    if err := req.Validate(); err != nil { return nil, err }
    if err := req.CheckPermission(ctx); err != nil { return nil, err }
    // business logic...
}
```

**Ownership case** — model needed for permission check (e.g. "is this actor the author?"):
```go
func UpdatePost(ctx context.Context, repo post.PostRepository, req post.UpdatePostRequest) (*post.PostView, error) {
    if err := req.Validate(); err != nil { return nil, err }

    existing, err := repo.FindByID(ctx, req.ID)
    if err != nil { ... }
    if existing == nil { return nil, post.ErrPostNotFound }

    if err := req.CheckPermission(ctx, existing); err != nil { return nil, err }
    // mutate and save...
}
```

The `CheckPermission` signature varies by request type — some take only `ctx`, others take `ctx` + the domain model. The domain model owns the authorization rule (e.g. `post.CanModify(actor)`) and `CheckPermission` delegates to it.

### 2. Services only accessible through the module API

Services must only be called through the module's exported API interface (e.g. `post.PostAPI`). No external code should import a service package directly to call it.

❌ Wrong:
```go
import postservice "github.com/noueii/no-frame-works/internal/modules/post/service"

// Calling service directly from outside the module
svc := postservice.New(repo)
svc.CreatePost(ctx, req)
```

✅ Correct:
```go
import "github.com/noueii/no-frame-works/internal/modules/post"

// Call through the module's API interface
var api post.PostAPI
api.CreatePost(ctx, req)
```

### 3. Request structs own Validate() and CheckPermission()

Every request struct must implement `Validate() error` and `CheckPermission(...) error`. These are defined on the request type in the module's `api.go` file, not in the service.

`CheckPermission` signature varies by case:
- Simple (actor/role check only): `CheckPermission(ctx context.Context) error`
- Ownership check: `CheckPermission(ctx context.Context, model *domain.Model) error`

The domain model owns authorization rules as pure methods (e.g. `CanModify(actor.Actor) bool`). `CheckPermission` calls these domain methods.

```go
// api.go — simple permission
func (r CreatePostRequest) CheckPermission(ctx context.Context) error {
    a := actor.ActorFrom(ctx)
    if a == nil { return ErrUnauthorized }
    return nil
}

// api.go — ownership permission, delegates to domain
func (r UpdatePostRequest) CheckPermission(ctx context.Context, post *domain.Post) error {
    a := actor.ActorFrom(ctx)
    if a == nil { return ErrUnauthorized }
    if !post.CanModify(a) { return ErrForbidden }
    return nil
}

// domain/post.go — pure business rule
func (p Post) CanModify(a actor.Actor) bool {
    if a.IsSystem() { return true }
    if ua, ok := a.(actor.UserActor); ok && ua.HasRole(actor.RoleAdmin) { return true }
    return p.AuthorID == a.UserID().String()
}
```

### 4. One function per file in service subfolders

Each file in a service subfolder (e.g. `service/create_post/create_post.go`) must contain exactly one exported function named after the operation (not `Execute`). The root `service/service.go` is the only file that can have multiple methods.

❌ Wrong:
```go
// service/create_post/create_post.go
func Execute(...) { ... }  // generic name — use the operation name
```

✅ Correct:
```go
// service/create_post/create_post.go
func CreatePost(ctx context.Context, repo post.PostRepository, req post.CreatePostRequest) (*post.PostView, error) {
    // single responsibility: create a post
}
```

### 5. Services use domain types internally

Service functions work with domain models from the module's `domain/` package for internal logic. They accept request structs as input and return view types as output — never domain models.

❌ Wrong:
```go
// Returning a domain model to the caller
func CreatePost(ctx context.Context, repo post.PostRepository, req post.CreatePostRequest) (*domain.Post, error) {
    // domain types should not leak outside the service
}
```

✅ Correct:
```go
func CreatePost(ctx context.Context, repo post.PostRepository, req post.CreatePostRequest) (*post.PostView, error) {
    // ...
    return &post.PostView{
        ID:       created.ID,
        Title:    created.Title,
        Content:  created.Content,
        AuthorID: created.AuthorID,
    }, nil
}
```

Service functions return pointers to view types. On error paths, return `nil, err` — never empty structs.

### 6. Constructor injection

Service structs receive their dependencies (repositories, providers) through the constructor. They never create dependencies internally.

❌ Wrong:
```go
func (s *Service) CreatePost(ctx context.Context, req post.CreatePostRequest) (post.PostView, error) {
    repo := postrepo.New(s.db)  // creating dependency inside method
    return createpost.Execute(ctx, repo, req)
}
```

✅ Correct:
```go
func New(repo post.PostRepository) *Service {
    return &Service{repo: repo}
}

func (s *Service) CreatePost(ctx context.Context, req post.CreatePostRequest) (post.PostView, error) {
    return createpost.Execute(ctx, s.repo, req)
}
```

### 7. Domain model is the source of truth

All operations go through the domain model. For updates: fetch the existing model from the repo, mutate its fields directly in memory, then send the complete model to the repo. The repository always receives a full domain model — never loose fields or partial objects.

For cross-module side effects (e.g. updating a user counter when a post is published), the service orchestrates by calling the other module's API. The domain model itself does not reach across modules.

❌ Wrong — bypassing the domain model with loose fields:
```go
func EditUsername(ctx context.Context, repo user.UserRepository, req user.EditUsernameRequest) (*user.UserView, error) {
    if err := req.Validate(); err != nil { ... }
    updated, err := repo.UpdateUsername(ctx, req.UserID, req.Username)
    return &user.UserView{ID: updated.ID, Username: updated.Username}, nil
}
```

✅ Correct:
```go
func UpdatePost(ctx context.Context, repo post.PostRepository, req post.UpdatePostRequest) (*post.PostView, error) {
    if err := req.Validate(); err != nil { return nil, err }

    existing, err := repo.FindByID(ctx, req.ID)
    if err != nil { ... }
    if existing == nil { return nil, post.ErrPostNotFound }

    if err := req.CheckPermission(ctx, existing); err != nil { return nil, err }

    existing.Title = req.Title
    existing.Content = req.Content

    updated, err := repo.Update(ctx, *existing)
    if err != nil { ... }

    return &post.PostView{
        ID:       updated.ID,
        Title:    updated.Title,
        Content:  updated.Content,
        AuthorID: updated.AuthorID,
    }, nil
}
```

Note: This applies to updates. For creates, the service builds a new domain model from the request fields since no existing model exists yet.

## Output Format

Only flag violations where you are at least 80% confident. Skip rules that don't apply to the diff. When in doubt, don't flag it.

For each violation, provide:
- Rule name
- File path
- The problematic code or function
- Brief explanation of what's wrong
