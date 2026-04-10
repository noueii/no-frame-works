# Repository Layer Review Rubric

You are reviewing repository code in a Go backend that uses go-jet as its type-safe query builder. Repositories live in `backend/repository/<module>/` and implement the module's `Repository` interface.

Repositories are **pure data access**. They translate between domain models and the database using go-jet. No business logic, no raw SQL, no manual column specifications.

## Allowed types

Repositories may only work with:
- **Domain models** (`domain.*`) — received from and returned to services
- **Database/go-jet models** (`model.*`) — for database operations

Repositories must NOT import or use:
- oapi-codegen generated types (`oapi.*`)
- API contract types (request structs, view types from `api.go`)

## Rules

### 1. No raw SQL queries

All database queries must use go-jet's type-safe query builder. No raw SQL strings passed to `db.Query`, `db.QueryRow`, `db.Exec`, or `db.QueryContext` and variants.

❌ Wrong:
```go
func (p *Postgres) FindByID(ctx context.Context, id string) (*domain.User, error) {
    var u domain.User
    err := p.db.QueryRowContext(ctx,
        `SELECT id, username, email FROM users WHERE id = $1`, id,
    ).Scan(&u.ID, &u.Username, &u.Email)
    return &u, err
}
```

❌ Wrong:
```go
func (p *Postgres) UpdateUsername(ctx context.Context, id string, username string) error {
    _, err := p.db.ExecContext(ctx,
        `UPDATE users SET username = $1 WHERE id = $2`, username, id,
    )
    return err
}
```

✅ Correct:
```go
func (r *PostgresPostRepository) Create(ctx context.Context, p domain.Post) (*domain.Post, error) {
    stmt := table.Post.INSERT(
        table.Post.Title,
        table.Post.Content,
        table.Post.AuthorID,
    ).MODEL(toModel(p)).
        RETURNING(table.Post.AllColumns)

    var dest model.Post
    err := stmt.QueryContext(ctx, r.db, &dest)
    if err != nil {
        return nil, fmt.Errorf("insert post: %w", err)
    }
    return toDomain(dest), nil
}
```

### 2. Use MODEL() with MutableColumns for updates

Updates must use go-jet's `MODEL()` with `MutableColumns` to update the full row from the model. Do not list individual columns manually.

If there is an edge case where `MutableColumns` cannot be used (e.g. a database-managed column that must not be overwritten), the code must include a comment explaining why and which columns are excluded.

❌ Wrong — manual column-value pairs:
```go
stmt := table.User.UPDATE().
    SET(table.User.Username.SET(String(username))).
    SET(table.User.UpdatedAt.SET(TimestampT(time.Now()))).
    WHERE(table.User.ID.EQ(String(id)))
```

❌ Wrong — listing columns individually without justification:
```go
stmt := table.User.UPDATE(
    table.User.Username,
    table.User.UpdatedAt,
).MODEL(updateModel).
    WHERE(table.User.ID.EQ(UUID(id)))
```

✅ Correct — MutableColumns updates the full row:
```go
stmt := table.User.UPDATE(table.User.MutableColumns).
    MODEL(toModel(u)).
    WHERE(table.User.ID.EQ(UUID(u.ID))).
    RETURNING(table.User.AllColumns)
```

✅ Also correct — explicit columns with a comment explaining the exception:
```go
// Using explicit columns because `rank` is computed by a database trigger
// and must not be overwritten on update.
stmt := table.User.UPDATE(
    table.User.Username,
    table.User.Email,
    table.User.UpdatedAt,
).MODEL(toModel(u)).
    WHERE(table.User.ID.EQ(UUID(u.ID))).
    RETURNING(table.User.AllColumns)
```

### 3. Mapping functions live in the repository

`toModel()` (domain → go-jet model) and `toDomain()` (go-jet model → domain) functions belong in the repository package, not in the domain or service layers.

❌ Wrong:
```go
// domain/models.go
func (p *Post) ToModel() model.Post { ... }

// service/create_post/create_post.go
func toModel(p domain.Post) model.Post { ... }
```

✅ Correct:
```go
// repository/post/create.go
func toModel(p domain.Post) model.Post {
    return model.Post{
        Title:    p.Title,
        Content:  p.Content,
        AuthorID: p.AuthorID,
    }
}

func toDomain(m model.Post) *domain.Post {
    return &domain.Post{
        ID:       m.ID.String(),
        Title:    m.Title,
        Content:  m.Content,
        AuthorID: m.AuthorID,
    }
}
```

### 4. No business logic in repositories

Repositories only do data access — querying, inserting, updating, deleting. No validation, no permission checks, no business rules, no conditional logic based on business state.

❌ Wrong:
```go
func (r *PostgresPostRepository) Create(ctx context.Context, p domain.Post) (*domain.Post, error) {
    if p.Title == "" {
        return nil, errors.New("title required")  // validation belongs in the service
    }
    if p.AuthorID == "" {
        return nil, errors.New("author required")
    }
    // ...
}
```

✅ Correct:
```go
func (r *PostgresPostRepository) Create(ctx context.Context, p domain.Post) (*domain.Post, error) {
    insert := toModel(p)
    stmt := table.Post.INSERT(...).MODEL(insert).RETURNING(table.Post.AllColumns)
    // just data access, no business rules
}
```

### 5. Model-in, model-out

Repository methods accept and return complete domain models. No narrow field-specific update methods. The service is responsible for deciding what changed — the repo just persists the full model it receives.

❌ Wrong — field-specific repo methods:
```go
// Repository interface with narrow update methods
type UserRepository interface {
    UpdateUsername(ctx context.Context, id string, username string) (*domain.User, error)
    UpdateEmail(ctx context.Context, id string, email string) (*domain.User, error)
    IncrementPostCount(ctx context.Context, id string) error
}
```

❌ Wrong — accepting loose fields:
```go
func (r *Repo) UpdateUsername(ctx context.Context, id string, username string) (*domain.User, error) {
    stmt := table.User.UPDATE(table.User.Username).
        MODEL(model.User{Username: username}).
        WHERE(table.User.ID.EQ(UUID(id)))
    // ...
}
```

✅ Correct — accept the full domain model:
```go
// Repository interface with simple CRUD operations
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*domain.User, error)
    Create(ctx context.Context, user domain.User) (*domain.User, error)
    Update(ctx context.Context, user domain.User) (*domain.User, error)
    Delete(ctx context.Context, id string) error
}
```

```go
func (r *Repo) Update(ctx context.Context, u domain.User) (*domain.User, error) {
    stmt := table.User.UPDATE(table.User.MutableColumns).
        MODEL(toModel(u)).
        WHERE(table.User.ID.EQ(UUID(u.ID))).
        RETURNING(table.User.AllColumns)

    var dest model.User
    err := stmt.QueryContext(ctx, r.db, &dest)
    if err != nil {
        return nil, fmt.Errorf("update user: %w", err)
    }
    return toDomain(dest), nil
}
```

### 6. One function per file in repository subfolders

Each file in a repository subfolder handles one operation. The root file (`postgres.go`) holds the struct definition, constructor, and compile-time interface check.

❌ Wrong:
```go
// repository/post/queries.go
func (r *PostgresPostRepository) FindByID(...) { ... }
func (r *PostgresPostRepository) ListAll(...) { ... }
func (r *PostgresPostRepository) Create(...) { ... }
```

✅ Correct:
```
repository/post/
├── postgres.go          # struct, New(), interface check
├── find_by_id.go        # FindByID
├── list_all.go          # ListAll
├── create.go            # Create + toModel/toDomain helpers
├── update.go            # Update
└── delete.go            # Delete
```

### 6. Provider pattern for external services

External SDKs and APIs must be accessed through provider interfaces, never called directly from repository or service code.

❌ Wrong:
```go
func (r *Repo) UploadAvatar(ctx context.Context, data []byte) (string, error) {
    client := s3.New(session.Must(session.NewSession()))
    _, err := client.PutObject(&s3.PutObjectInput{...})
}
```

✅ Correct:
```go
// Defined as a provider interface
type StorageProvider interface {
    Upload(ctx context.Context, key string, data []byte) (string, error)
}

// Injected into the service/repo that needs it
func New(db *sql.DB, storage StorageProvider) *Repo {
    return &Repo{db: db, storage: storage}
}
```

## Output Format

Only flag violations where you are at least 80% confident. Skip rules that don't apply to the diff. When in doubt, don't flag it.

For each violation, provide:
- Rule name
- File path
- The problematic code or function
- Brief explanation of what's wrong
