---
applies-to: ["*.go", "internal/app/services/**/*.go", "services/**/repository*.go"]
---

# Go Repository Layer Rules

## Structure
- `repository.go` in each service: repo struct + constructor
- One function per file in per-entity sub-packages: `find_by_id.go`, `create.go`, etc.

## Technology
- **Only go-jet** query builder - NO raw SQL
- Updates use `MODEL(MutableColumns)` with exceptions commented
- No business logic - only data access

## Mapping
- `toModel()`: domain → go-jet model
- `toDomain()`: go-jet model → domain
- Mapping functions are private to repository package

## Domain-in, Domain-out
- Repository receives complete domain model
- Returns domain model
- No partial updates - always full domain value

## Naming
- `FindByID(id string) (*domain.Post, error)`
- `Create(post *domain.Post) (*domain.Post, error)`
- `Update(post *domain.Post) (*domain.Post, error)`
- `Delete(id string) error`

## Example
```go
func (r *Repository) FindByID(ctx context.Context, id string) (*domain.Post, error) {
    row := posttable.FindById(id)
    return toDomain(row), nil
}

func toDomain(row posttable.PostTableType) *domain.Post {
    return &domain.Post{
        ID: domain.PostID(row.ID),
        Title: row.Title,
        // ...
    }
}
```