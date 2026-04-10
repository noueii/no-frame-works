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

### 1. Validate first

Every service method (or its delegated Execute function) must call `req.Validate()` as its first meaningful operation. Nothing else should happen before validation.

❌ Wrong:
```go
func Execute(ctx context.Context, repo user.UserRepository, req user.EditUsernameRequest) (user.UserView, error) {
    existing, err := repo.FindByID(ctx, req.UserID)  // repo call before validation
    if err != nil {
        return user.UserView{}, err
    }
    if err := req.Validate(); err != nil {
        return user.UserView{}, err
    }
}
```

✅ Correct:
```go
func Execute(ctx context.Context, repo user.UserRepository, req user.EditUsernameRequest) (user.UserView, error) {
    if err := req.Validate(); err != nil {
        return user.UserView{}, fmt.Errorf("validation failed: %w", err)
    }
    existing, err := repo.FindByID(ctx, req.UserID)
    // ...
}
```

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

### 3. Request structs own Validate() and Permission()

Every request struct must implement both `Validate() error` and `Permission() Permission` methods. These are defined on the request type in the module's `api.go` file, not in the service.

❌ Wrong:
```go
// Validation inside the service instead of on the request
func Execute(ctx context.Context, repo user.UserRepository, req user.EditUsernameRequest) (user.UserView, error) {
    if req.Username == "" {
        return user.UserView{}, errors.New("username required")
    }
}
```

✅ Correct:
```go
// Validation on the request struct in api.go
func (r EditUsernameRequest) Validate() error {
    if r.Username == "" {
        return ErrUsernameRequired
    }
    return nil
}

func (r EditUsernameRequest) Permission() Permission {
    return PermUserEdit
}

// Service just calls req.Validate()
func Execute(ctx context.Context, repo user.UserRepository, req user.EditUsernameRequest) (user.UserView, error) {
    if err := req.Validate(); err != nil {
        return user.UserView{}, fmt.Errorf("validation failed: %w", err)
    }
}
```

### 4. One function per file in service subfolders

Each file in a service subfolder (e.g. `service/create_post/create_post.go`) must contain exactly one exported function: `Execute`. The root `service/service.go` is the only file that can have multiple methods.

❌ Wrong:
```go
// service/create_post/create_post.go
func Execute(...) { ... }
func validateTitle(...) { ... }  // helper should be in the same file but not exported, OR logic belongs in Validate()
func NotifyAuthor(...) { ... }   // second exported function — must be in its own subfolder
```

✅ Correct:
```go
// service/create_post/create_post.go
func Execute(ctx context.Context, repo post.PostRepository, req post.CreatePostRequest) (post.PostView, error) {
    // single responsibility: create a post
}
```

### 5. Services use domain types internally

Service functions work with domain models from the module's `domain/` package for internal logic. They accept request structs as input and return view types as output — never domain models.

❌ Wrong:
```go
// Returning a domain model to the caller
func Execute(ctx context.Context, repo post.PostRepository, req post.CreatePostRequest) (*domain.Post, error) {
    // domain types should not leak outside the service
}
```

✅ Correct:
```go
func Execute(ctx context.Context, repo post.PostRepository, req post.CreatePostRequest) (post.PostView, error) {
    // ...
    return post.PostView{
        ID:       created.ID,
        Title:    created.Title,
        Content:  created.Content,
        AuthorID: created.AuthorID,
    }, nil
}
```

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
func Execute(ctx context.Context, repo user.UserRepository, req user.EditUsernameRequest) (user.UserView, error) {
    if err := req.Validate(); err != nil { ... }
    // Never loaded the domain model — just sent raw fields to repo
    updated, err := repo.UpdateUsername(ctx, req.UserID, req.Username)
    return user.UserView{ID: updated.ID, Username: updated.Username}, nil
}
```

❌ Wrong — constructing a model from request fields instead of fetching:
```go
func Execute(ctx context.Context, repo post.PostRepository, req post.UpdatePostRequest) (post.PostView, error) {
    if err := req.Validate(); err != nil { ... }
    // Built from scratch — doesn't reflect actual current state
    updated := domain.Post{ID: req.ID, Title: req.Title, Content: req.Content}
    result, err := repo.Update(ctx, updated)
}
```

✅ Correct:
```go
func Execute(ctx context.Context, repo post.PostRepository, req post.UpdatePostRequest) (post.PostView, error) {
    if err := req.Validate(); err != nil { ... }

    // 1. Fetch existing model — this is the source of truth
    existing, err := repo.FindByID(ctx, req.ID)
    if err != nil { ... }
    if existing == nil { return post.PostView{}, post.ErrPostNotFound }

    // 2. Mutate fields directly
    existing.Title = req.Title
    existing.Content = req.Content

    // 3. Send complete model to repository
    updated, err := repo.Update(ctx, *existing)
    if err != nil { ... }

    return post.PostView{
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
