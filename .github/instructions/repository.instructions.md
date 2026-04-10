---
applyTo: "backend/repository/**"
---

# Repository Layer Conventions

Repositories are pure data access using go-jet. They translate between domain models and the database.

## Allowed types
- `domain.*` models (received from/returned to services) and `model.*` (go-jet/database) types
- Must NOT use: `oapi.*` (OpenAPI generated) or API contract types (request structs, view types)

## Rules

1. **No raw SQL** — All queries must use go-jet's type-safe query builder. No raw SQL strings passed to `db.Query`, `db.QueryRow`, `db.Exec`, or their Context variants.

2. **MODEL() with MutableColumns for updates** — Updates use `MODEL()` with `MutableColumns`. If an edge case requires explicit columns, a comment must explain why.

3. **Mapping functions live here** — `toModel()` (domain → go-jet model) and `toDomain()` (go-jet model → domain) belong in the repository, not in domain or service.

4. **No business logic** — Only data access. No validation, no permission checks, no conditional business logic.

5. **Model-in, model-out** — Repository methods accept and return complete domain models. No narrow field-specific update methods like `UpdateUsername()` or `IncrementCount()`. Simple CRUD: `FindByID`, `Create`, `Update`, `Delete`.

6. **One function per file** — Each file handles one operation. Root file (`postgres.go`) holds struct, constructor, and interface check.

7. **Provider pattern for external services** — External SDKs behind provider interfaces, injected through constructors.
