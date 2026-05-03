# Request Flow Integrity Rubric

You are reviewing the end-to-end request lifecycle across all layers. Your job is NOT to review individual layers (other agents handle that). Your job is to trace the data flow from handler to repository and back, and flag any steps that are skipped, out of order, or incorrectly wired.

## The Expected Flow

Every request follows this lifecycle:

```
1. Handler receives oapi request object
2. Handler transforms oapi fields → api.Op struct with Request
3. Handler calls h.app.API().Posts.CreatePost(ctx, op)
4. Service validates inline: op.Request.Field == ""
5. Service performs business logic
6. Service populates state on op (e.g. op.Post = result)
7. Service calls s.app.API().OtherService.Method() if cross-service call needed
8. Service calls repository with domain model
9. Repository converts domain model → go-jet model (toModel)
10. Repository executes go-jet query
11. Repository converts go-jet model → domain model (toDomain)
12. Repository returns domain model to service
13. Service returns domain model to handler (state available on op)
14. Handler converts domain model → oapi response type
15. Handler returns oapi response
```

## Rules

### 1. No skipped steps

Trace the flow for each new or modified endpoint. Every step in the lifecycle must be present.

❌ Wrong — handler calls repo directly (skips service):
```go
func (h *Handler) GetPost(ctx context.Context, req oapi.GetPostRequestObject) (oapi.GetPostResponseObject, error) {
    post, err := h.repo.FindByID(ctx, req.Id.String())  // skips service
}
```

❌ Wrong — service returns domain model without op state:
```go
func (s *Service) GetPost(ctx context.Context, op *api.GetPostOp) (*domain.Post, error) {
    return s.repo.FindByID(ctx, op.Request.ID)  // op.Post never set
}
```

### 2. Correct data type at each boundary

| Layer | Uses | Must NOT use |
|-------|------|-------------|
| Handler | `oapi.*` + `services/api` Op types | `domain.*`, `model.*` |
| Service | `services/api` Op types + `domain.*` | `oapi.*`, `model.*` |
| Repository | `domain.*` + `model.*` (go-jet) | `oapi.*`, API types |

Each boundary has a specific type:

| Boundary | Type crossing |
|----------|---------------|
| Handler → Service | `*api.CreatePostOp` with `Request` set |
| Service → Repository | `domain.Post` |
| Repository → Service | `domain.Post` |
| Service → Handler | `*domain.Post` (op state available) |
| Handler → Response | `oapi.*` |

❌ Wrong — handler passes wrong type:
```go
h.app.API().Posts.CreatePost(ctx, &api.CreatePostOp{
    Request: request.Body,  // should be api.CreatePostRequest
})
```

❌ Wrong — repository returns go-jet model:
```go
func (r *Repo) FindByID(...) (*model.Post, error) {
    return &dest, nil  // should return toDomain(&dest)
}
```

### 3. Op structure is respected

Op has two parts:
- `Request` — what caller provides (set by handler)
- State fields (e.g. `Post`) — set by service

Handlers only set `Request`. Services set state fields.

❌ Wrong — handler sets state:
```go
h.app.API().Posts.CreatePost(ctx, &api.CreatePostOp{
    Request: api.CreatePostRequest{...},
    Post:    somePost,  // handler should not set this
})
```

❌ Wrong — service doesn't populate state:
```go
func (s *Service) GetPost(ctx context.Context, op *api.GetPostOp) (*domain.Post, error) {
    post, _ := s.repo.FindByID(ctx, op.Request.ID)
    return post, nil  // op.Post never set, caller can't access state
}
```

### 4. Cross-service calls use App.API()

When service A calls service B:

```go
// Service A
s.app.API().Users.IncrementPostCount(ctx, &api.IncrementPostCountOp{
    Request: api.IncrementPostCountRequest{UserID: op.Request.AuthorID},
})
```

Never call another service's repository directly.

### 5. Return path is complete

State flows back through the Op. Domain model is returned directly (not wrapped in a view type in this architecture).

```go
// service
func (s *Service) GetPost(ctx context.Context, op *api.GetPostOp) (*domain.Post, error) {
    post, _ := s.repo.FindByID(ctx, op.Request.ID)
    op.Post = post  // state available to handler
    return post, nil
}

// handler
result, err := h.app.API().Posts.GetPost(ctx, &api.GetPostOp{
    Request: api.GetPostRequest{ID: request.Id.String()},
})
// result is *domain.Post, op.Post also available if needed
```

## File Structure

```
services/
├── api/
│   ├── post.go    # Request types + Op types + PostAPI interface
│   └── user.go    # Request types + Op types + UserAPI interface
├── post/
│   ├── api.go           # Service struct, implements api.PostAPI
│   ├── create_post.go   # CreatePost method
│   ├── get_post.go      # GetPost method
│   └── ...
└── user/
    ├── api.go
    ├── get_user.go
    └── ...
```

## Output Format

Only flag violations where you are at least 80% confident. Skip endpoints where the flow is correct. When in doubt, don't flag it.

For each violation, provide:
- Which step in the flow is broken
- File paths involved
- What's wrong and what the correct flow should be