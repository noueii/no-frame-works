# Service Layer Review Rubric

You are reviewing service code in a Go backend with a modular architecture.

Services are accessed exclusively through the module's API interface in `services/api/`. Each service implements its interface and holds business logic.

## Allowed types

Services may only work with:
- **API Op types** from `services/api/` (e.g. `*api.CreatePostOp`)
- **Domain models** (`domain.*`) — for internal business logic

Services must NOT import or use:
- oapi-codegen generated types (`oapi.*`)
- Database/go-jet models (`model.*`)

## Rules

### 1. Op holds Request + State

Ops live in `services/api/` and have two parts:

```go
// services/api/post.go
type CreatePostOp struct {
    Request CreatePostRequest  // input fields
    Post    *domain.Post        // state populated during execution
}
```

Service methods receive `*api.CreatePostOp`. Validation uses `op.Request.Field`. State is stored in op fields like `op.Post`.

### 2. Inline validation, no Op.Validate()

Validation is done inline in the service method using `op.Request.Field == ""` checks. No `Validate()` method on Op.

❌ Wrong:
```go
if err := op.Validate(); err != nil { return nil, err }
```

✅ Correct:
```go
if op.Request.Title == "" {
    return nil, apperrors.Validation(apperrors.CodePostTitleRequired, "title is required", nil)
}
```

### 3. One service method per file in service subfolders

Each operation has its own file. The `api.go` defines the `Service` struct and implements `api.PostAPI`. Each `*_post.go` file contains one service method.

```go
// services/post/api.go
type Service struct {
    app  *config.App
    repo PostRepository
}
func (s *Service) CreatePost(ctx context.Context, op *api.CreatePostOp) (*domain.Post, error)

// services/post/create_post.go
func (s *Service) CreatePost(ctx context.Context, op *api.CreatePostOp) (*domain.Post, error) {
    if op.Request.Title == "" { ... }
    // ...
}
```

### 4. Cross-service calls through App.API()

When a service needs to call another service, it uses `s.app.API().Users.IncrementPostCount(...)`.

❌ Wrong — direct repo access to another service's data:
```go
s.userRepo.IncrementPostCount(...)
```

✅ Correct — through the API interface:
```go
s.app.API().Users.IncrementPostCount(ctx, &api.IncrementPostCountOp{
    Request: api.IncrementPostCountRequest{UserID: op.Request.AuthorID},
})
```

### 5. State flows through Op

Service populates state fields on the Op so the caller can access them:

```go
func (s *Service) GetPost(ctx context.Context, op *api.GetPostOp) (*domain.Post, error) {
    post, err := s.repo.FindByID(ctx, op.Request.ID)
    // ...
    op.Post = post  // caller can access op.Post
    return post, nil
}
```

### 6. Constructor injection

Service struct receives dependencies through the constructor:

```go
func New(app *config.App, repo PostRepository) *Service {
    return &Service{app: app, repo: repo}
}
```

Never create dependencies inside service methods.

### 7. Domain model is source of truth for updates

For updates: fetch existing domain model → mutate fields → save complete model.

```go
func (s *Service) UpdatePost(ctx context.Context, op *api.UpdatePostOp) (*domain.Post, error) {
    existing, err := s.repo.FindByID(ctx, op.Request.ID)
    if err != nil { return nil, err }
    
    op.Post = existing  // store state
    op.Post.Title = op.Request.Title
    op.Post.Content = op.Request.Content
    
    updated, err := s.repo.Update(ctx, *op.Post)
    return updated, nil
}
```

## API Structure

```go
// services/api/post.go
type CreatePostRequest struct {
    Title    string
    Content  string
    AuthorID string
}

type CreatePostOp struct {
    Request CreatePostRequest
}

type PostAPI interface {
    CreatePost(ctx context.Context, op *CreatePostOp) (*domain.Post, error)
    GetPost(ctx context.Context, op *GetPostOp) (*domain.Post, error)
    // ...
}

// services/post/api.go
type Service struct {
    app  *config.App
    repo PostRepository
}
var _ api.PostAPI = (*Service)(nil)
```

## Output Format

Only flag violations where you are at least 80% confident. Skip rules that don't apply to the diff. When in doubt, don't flag it.

For each violation, provide:
- Rule name
- File path
- The problematic code or function
- Brief explanation of what's wrong