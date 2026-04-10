# Code Review Instructions

When reviewing pull requests, check for these conventions.

## Request Flow Integrity

Every request follows this lifecycle:
1. Handler receives oapi request → transforms to service request struct
2. Service calls `req.Validate()` first → performs business logic with domain models
3. Repository receives complete domain model → converts to go-jet model → executes query
4. Return path: repo returns domain model → service converts to view type → handler converts to oapi response

Flag any skipped steps, wrong types at boundaries, or layer bypassing (e.g. handler importing repository, service calling database directly).

## Common Mistake Patterns

1. **Use `errors.New()` for static errors, `fmt.Errorf()` only for wrapping** — `fmt.Errorf("user not found")` is wrong, should be `errors.New()` or a sentinel error.

2. **Always wrap with `%w`, not `%v` or `%s`** — Preserves the error chain for `errors.Is()` and `errors.As()`.

3. **Return sentinel errors directly** — Don't wrap domain errors like `ErrNotFound` with `fmt.Errorf`. Return them as-is. Only wrap infrastructure errors from repos/providers.

## Type Isolation

Each layer has strict type boundaries:
- **Handler**: uses `oapi.*` + API contract types (`Request`, `View`). No `domain.*` or `model.*`.
- **Service**: uses API contract types + `domain.*`. No `oapi.*` or `model.*`.
- **Repository**: uses `domain.*` + `model.*` (go-jet). No `oapi.*` or API contract types.

## Architecture Rules

- Handlers only transform data, never validate or contain business logic
- Services are the only place for business logic
- Domain models are the source of truth — for updates, always fetch → mutate → send complete model
- Repositories use go-jet only, no raw SQL, updates use `MutableColumns`
- Repository methods are simple CRUD accepting complete domain models
- Each module owns its types — no shared domain packages
- Dependencies injected through constructors, never created internally
