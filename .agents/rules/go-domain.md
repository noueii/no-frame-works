---
applies-to: ["*.go", "internal/app/domain/**/*.go", "services/**/domain/**/*.go"]
---

# Go Domain Layer Rules

## Structure
- Domain types in `internal/app/domain/<entity>/` or `services/<name>/domain/`
- One file per domain type: `post.go`, `user.go`
- `errors.go` in domain root for sentinel errors
- Related logic in separate files: `post_permissions.go`, `post_validation.go`

## Naming
- Domain types: `Post`, `User`, `PostID` (value object)
- No prefix like `Domain` or `Entity`
- Value objects use suffix: `PostID`, `UserID`, `Email`

## Rules
- **Pure business logic only**
- No imports: `database/sql`, `net/http`, external SDKs
- No persistence concerns (no `ToJSON`, `ToDBModel`)
- Authorization as methods: `CanModify(actor UserID) bool`
- Validation on root aggregates: `Validate() error`

## Sentinel Errors
```go
// internal/app/apperrors/errors.go
var ErrNotFound = errors.New("not found")
var ErrValidation = errors.New("validation failed")
```

## Errors
- Define sentinels in `domain/errors.go` (not inline, not in module root)
- Wrap at every layer: `errors.Errorf("domain.post.Create: %w", err)`

## Domain Model Example
```go
type Post struct {
    ID        PostID
    Title     string
    AuthorID  UserID
    CreatedAt time.Time
}

func (p *Post) CanModify(actor UserID) bool {
    return p.AuthorID == actor
}

func (p *Post) Validate() error {
    if p.Title == "" {
        return ErrTitleRequired
    }
    return nil
}
```