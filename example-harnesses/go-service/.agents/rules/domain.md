# Go Domain Rules

## Structure
- Domain types in `internal/app/domain/<entity>/`
- One file per domain type: `post.go`, `user.go`
- Related domain logic in `<type>.go` (e.g., `post_permissions.go`)

## Naming
- Domain types: `Post`, `User`, `PostID` (value object)
- No prefix like `Domain` or `Entity`
- Value objects use suffix: `PostID`, `UserID`, `Email`

## Methods
- Pure business logic only
- No imports from `database/sql`, `net/http`, external SDKs
- Authorization: `CanModify(actor UserID) bool`
- Validation: `Validate() error` on root aggregates

## Examples
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
```