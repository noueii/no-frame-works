# Domain Layer Review Rubric

You are reviewing domain code in a Go backend. Domain types are consolidated in `internal/app/domain/`, not per-service.

The domain layer is the **heart of the module**. It defines the data structures and business rules. It has zero knowledge of infrastructure (databases, HTTP, external services).

## Rules

### 1. No infrastructure imports

Domain packages must not import any infrastructure packages — no database drivers, HTTP libraries, external SDKs, or framework code.

❌ Wrong:
```go
package domain

import (
    "database/sql"
    "net/http"
    "github.com/lib/pq"
)
```

✅ Correct:
```go
package domain

import (
    "time"
    "errors"
)
```

### 2. Domain models are pure data + business logic

Domain structs represent business entities. Methods on domain types must be pure business logic — computations, state checks, transformations. No I/O, no database calls, no HTTP calls.

❌ Wrong:
```go
func (p *Post) Save(db *sql.DB) error { ... }
func (u *User) FetchProfile(client *http.Client) error { ... }
```

✅ Correct:
```go
func (p *Post) IsOwnedBy(authorID string) bool {
    return p.AuthorID == authorID
}

func (u *User) CanEdit() bool {
    return !u.IsLocked
}
```

### 3. User-facing errors use apperrors package

Domain-specific sentinel errors live in `internal/app/core/apperrors/`, not in domain package. Use `apperrors.NotFound/Validation/Conflict/etc.` for user-facing errors.

❌ Wrong — sentinel in domain:
```go
var ErrPostNotFound = errors.New("not found")
```

✅ Correct — use shared apperrors:
```go
apperrors.NotFound(apperrors.CodePostNotFound, "post not found", map[string]any{"post_id": id})
```

### 4. Types in consolidated domain package

All domain types live in `internal/app/domain/`, not per-service. Services reference domain types, domain types don't reference other service types (use IDs for cross-service references).

❌ Wrong:
```go
type Post struct {
    Author user.User  // importing another service's type
}
```

✅ Correct:
```go
type Post struct {
    AuthorID string  // reference by ID
}
```

### 5. Domain functions must be business logic

If a function on a domain type doesn't express a business rule or business computation, it doesn't belong in the domain. Mapping functions (`toModel`, `toDomain`) belong in the repository.

❌ Wrong:
```go
func (p *Post) ToDBModel() model.Post { ... }  // repository concern
```

✅ Correct:
```go
func (p *Post) IsOwnedBy(authorID string) bool { ... }  // business logic
```

## Output Format

Only flag violations where you are at least 80% confident. Skip rules that don't apply to the diff. When in doubt, don't flag it.

For each violation, provide:
- Rule name
- File path
- The problematic code or function
- Brief explanation of what's wrong