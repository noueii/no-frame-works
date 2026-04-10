# Domain Layer Review Rubric

You are reviewing domain code in a Go backend with a modular architecture. Each module has a `domain/` package that contains the module's internal models and business logic.

The domain layer is the **heart of the module**. It defines the data structures and business rules. It has zero knowledge of infrastructure (databases, HTTP, external services).

## Rules

### 1. No infrastructure imports

Domain packages must not import any infrastructure packages — no database drivers, HTTP libraries, external SDKs, or framework code. Only standard library types (e.g. `time`, `errors`, `fmt`, `strings`) and other domain packages within the same module.

❌ Wrong:
```go
package domain

import (
    "database/sql"
    "net/http"
    "github.com/lib/pq"
    "github.com/go-chi/chi/v5"
)
```

✅ Correct:
```go
package domain

import (
    "time"
    "errors"
    "fmt"
)
```

### 2. Domain models are pure data + business logic

Domain structs represent business entities. Methods on domain types must be pure business logic — computations, state checks, transformations. No I/O, no database calls, no HTTP calls.

❌ Wrong:
```go
func (p *Post) Save(db *sql.DB) error {
    _, err := db.Exec("INSERT INTO posts ...")
    return err
}

func (u *User) FetchProfile(client *http.Client) error {
    resp, _ := client.Get("https://api.example.com/...")
    // ...
}
```

✅ Correct:
```go
func (p *Post) IsOwnedBy(authorID string) bool {
    return p.AuthorID == authorID
}

func (u *User) CanEditUsername() bool {
    return u.Username != "" && !u.IsLocked
}
```

### 3. Sentinel errors in domain/errors.go

Module-specific domain errors must be defined as sentinel errors in the module's `domain/errors.go` file. Not inline with `fmt.Errorf`, not in the module root, not in service files.

❌ Wrong — errors in module root:
```go
// internal/modules/user/errors.go
package user

var ErrUserNotFound = errors.New("user not found")
```

❌ Wrong — errors created inline in service:
```go
func Execute(...) {
    if existing == nil {
        return user.UserView{}, fmt.Errorf("user not found")  // should be a sentinel error
    }
}
```

✅ Correct:
```go
// internal/modules/user/domain/errors.go
package domain

import "errors"

var (
    ErrUserNotFound  = errors.New("user not found")
    ErrUsernameTaken = errors.New("username is already taken")
)
```

Note: `fmt.Errorf` wrapping of existing sentinel errors is fine in services (e.g. `fmt.Errorf("failed to find user: %w", err)`). Only flag `fmt.Errorf` that creates new domain error concepts.

### 4. Types owned by their module

Domain types live in the module that owns them. No shared domain packages. Modules must not import another module's `domain/` package.

❌ Wrong:
```go
package domain

import (
    userdomain "github.com/noueii/no-frame-works/internal/modules/user/domain"
)

type Post struct {
    Author userdomain.User  // importing another module's domain
}
```

✅ Correct:
```go
package domain

type Post struct {
    AuthorID string  // reference by ID, not by importing the other module's type
}
```

### 5. Domain functions must be business logic

If a function on a domain type doesn't express a business rule or business computation, it probably doesn't belong in the domain. Mapping functions (`toModel`, `toDomain`) belong in the repository. Formatting functions belong in the handler or a presentation layer.

❌ Wrong:
```go
// domain/models.go
func (p *Post) ToJSON() ([]byte, error) {
    return json.Marshal(p)  // presentation concern
}

func (p *Post) ToDBModel() model.Post {
    return model.Post{Title: p.Title}  // persistence concern
}
```

✅ Correct:
```go
// domain/models.go
func (p *Post) IsPublished() bool {
    return p.PublishedAt != nil && p.PublishedAt.Before(time.Now())
}

// repository/post/create.go — mapping lives here
func toModel(p domain.Post) model.Post {
    return model.Post{Title: p.Title, Content: p.Content}
}
```

## Output Format

Only flag violations where you are at least 80% confident. Skip rules that don't apply to the diff. When in doubt, don't flag it.

For each violation, provide:
- Rule name
- File path
- The problematic code or function
- Brief explanation of what's wrong
