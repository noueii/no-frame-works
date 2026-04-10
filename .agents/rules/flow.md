# Request Flow Integrity Rubric

You are reviewing the end-to-end request lifecycle across all layers. Your job is NOT to review individual layers (other agents handle that). Your job is to trace the data flow from handler to repository and back, and flag any steps that are skipped, out of order, or incorrectly wired.

## The Expected Flow

Every request should follow this lifecycle in order:

```
1. Handler receives oapi request object
2. Handler transforms oapi fields → service request struct
3. Handler calls the module's API interface
4. Service calls req.Validate() FIRST
5. Service performs business logic (using domain models from repo)
6. Service calls repository methods with domain models
7. Repository converts domain model → go-jet model (toModel)
8. Repository executes go-jet query
9. Repository converts go-jet model → domain model (toDomain)
10. Repository returns domain model to service
11. Service converts domain model → view type
12. Service returns view type to handler
13. Handler converts view type → oapi response type
14. Handler returns oapi response
```

## Rules

### 1. No skipped steps

Trace the flow for each new or modified endpoint. Every step in the lifecycle must be present. Flag if any step is missing.

❌ Wrong — handler calls repo directly (skips service):
```go
// handler
func (h *Handler) GetPost(ctx context.Context, req oapi.GetPostRequestObject) (oapi.GetPostResponseObject, error) {
    post, err := h.repo.FindByID(ctx, req.Id.String())  // skips service entirely
    return oapi.GetPost200JSONResponse{...}, nil
}
```

❌ Wrong — service skips validation:
```go
// service
func Execute(ctx context.Context, repo post.PostRepository, req post.CreatePostRequest) (post.PostView, error) {
    newPost := domain.Post{Title: req.Title}  // no req.Validate() call
    created, err := repo.Create(ctx, newPost)
}
```

❌ Wrong — service returns domain model instead of view type:
```go
func Execute(...) (*domain.Post, error) {
    return repo.FindByID(ctx, req.ID)  // leaks domain model out of service
}
```

### 2. Correct data type at each boundary

Each layer has strict type isolation — types must not leak across boundaries:

| Layer | Uses | Must NOT use |
|-------|------|-------------|
| Handler | `oapi.*` + API contract (`Request`, `View`) | `domain.*`, `model.*` |
| Service | API contract + `domain.*` | `oapi.*`, `model.*` |
| Repository | `domain.*` + `model.*` (go-jet) | `oapi.*`, API contract |

Each boundary has a specific type that should cross it:

| Boundary | Type crossing it |
|----------|-----------------|
| Handler → Service | API contract request struct (e.g. `post.CreatePostRequest`) |
| Service → Repository | Domain model (e.g. `domain.Post`) |
| Repository → Database | go-jet model (e.g. `model.Post`) |
| Database → Repository | go-jet model → converted to domain model |
| Repository → Service | Domain model |
| Service → Handler | API contract view type (e.g. `post.PostView`) |
| Handler → Response | oapi response type |

❌ Wrong — passing oapi types to service:
```go
// handler passes raw oapi body to service
result, err := h.postAPI.CreatePost(ctx, request.Body)
```

❌ Wrong — passing go-jet model to service:
```go
// repository returns go-jet model instead of domain model
func (r *Repo) FindByID(ctx context.Context, id string) (*model.Post, error) {
    var dest model.Post
    // ...
    return &dest, nil  // should return toDomain(dest)
}
```

✅ Correct — each boundary has the right type:
```go
// handler: oapi → request struct
result, err := h.postAPI.CreatePost(ctx, post.CreatePostRequest{
    Title:   request.Body.Title,
    Content: request.Body.Content,
})

// service: request struct → domain model → view type
func Execute(ctx context.Context, repo post.PostRepository, req post.CreatePostRequest) (post.PostView, error) {
    if err := req.Validate(); err != nil { ... }
    newPost := domain.Post{Title: req.Title, Content: req.Content}
    created, err := repo.Create(ctx, newPost)
    return post.PostView{ID: created.ID, Title: created.Title}, nil
}

// repository: domain model → go-jet model → domain model
func (r *Repo) Create(ctx context.Context, p domain.Post) (*domain.Post, error) {
    stmt := table.Post.INSERT(...).MODEL(toModel(p)).RETURNING(table.Post.AllColumns)
    var dest model.Post
    err := stmt.QueryContext(ctx, r.db, &dest)
    return toDomain(dest), nil
}
```

### 3. No layer bypassing

Each layer must go through the next layer in sequence. No skipping.

❌ Wrong — handler imports repository:
```go
import postrepo "github.com/noueii/no-frame-works/repository/post"
```

❌ Wrong — handler imports domain:
```go
import "github.com/noueii/no-frame-works/internal/modules/post/domain"
```

❌ Wrong — service calls database directly:
```go
func Execute(ctx context.Context, db *sql.DB, req post.CreatePostRequest) (post.PostView, error) {
    db.ExecContext(ctx, "INSERT INTO posts ...")  // service bypasses repository
}
```

### 4. Return path is complete

The return path must mirror the request path. Data coming back from the database must be converted at each boundary, not passed through raw.

❌ Wrong — repo result passed through without conversion:
```go
// service returns repo result without converting to view
func (s *Service) GetPost(ctx context.Context, req post.GetPostRequest) (*domain.Post, error) {
    return s.repo.FindByID(ctx, req.ID)  // returns domain model, should be PostView
}
```

❌ Wrong — handler returns service result without converting to oapi type:
```go
func (h *Handler) GetPost(ctx context.Context, req oapi.GetPostRequestObject) (oapi.GetPostResponseObject, error) {
    result, _ := h.postAPI.GetPost(ctx, ...)
    return result, nil  // result is PostView, not an oapi response type
}
```

### 5. New endpoints are fully wired

When a new API method is added, verify the complete chain exists:
- [ ] oapi handler method implemented
- [ ] Service request struct defined with Validate() and Permission()
- [ ] Service function/Execute created
- [ ] Repository method exists if data access is needed
- [ ] Module API interface includes the new method
- [ ] All type conversions are present at each boundary

## How To Review

1. Identify new or modified endpoints in the diff
2. For each endpoint, trace the flow from handler → service → repository and back
3. Check that each step exists and uses the correct types
4. Flag any missing steps, wrong types at boundaries, or layer bypassing

## Output Format

Only flag violations where you are at least 80% confident. Skip endpoints where the flow is correct. When in doubt, don't flag it.

For each violation, provide:
- Which step in the flow is broken
- File paths involved
- What's wrong and what the correct flow should be
