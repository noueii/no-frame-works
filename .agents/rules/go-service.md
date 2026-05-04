---
applies-to: ["*.go", "internal/app/services/**/*.go", "services/**/service*.go"]
---

# Go Service Layer Rules

## Structure
- One file per operation: `create_post.go`, `get_post.go`, etc.
- `api.go` in service root: Request/Op types, API interface, Validate()
- `service.go` in each service: struct + constructor + business logic

## Types
- **Request structs**: Input data, e.g. `CreatePostRequest`
- **Op structs**: `Op{ Request, state fields }`. Request = input, Op holds mutable state
- **View structs**: Output/response structs, returned as `*View`
- **API interface**: Exported interface per service (what handlers use)

## Validation
- `req.Validate()` called FIRST in every service method
- `req.CheckPermission(actor)` called SECOND (if applicable)
- Inline validation, no Validate() method on Op types
- Return typed errors from `apperrors` package

## Op Pattern
```go
type CreatePostOp struct {
    Request CreatePostRequest
    // state fields populated by service
}

func (s *Service) CreatePost(ctx context.Context, op *api.CreatePostOp) (*domain.Post, error) {
    if op.Request.Title == "" {
        return nil, apperrors.Validation(...)
    }
    // business logic
}
```

## Repository Access
- Service calls repo methods only
- Repo returns domain models
- Service converts domain → View pointer

## Errors
- Wrap at every layer: `errors.Errorf("service.post.CreatePost: %w", err)`
- Use `github.com/go-errors/errors` (never `fmt.Errorf`)
- Return `nil, err` on failure (never empty structs)