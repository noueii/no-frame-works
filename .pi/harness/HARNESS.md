---
name: go-backend-harness
description: >
  Harness for the no-frame-works Go backend project. Enforces layered architecture
  (handler/service/repository/domain), type isolation, and error handling conventions.
  Activate when working on backend Go code, writing handlers, services, repositories,
  or domain models. Use /harness-review to run a full architecture review.
---

# Go Backend Harness

This harness enforces the project's strict layered architecture. All code must follow
the rules defined in `.agents/rules/`.

## Quick Reference

### Type Isolation (never cross layers)
| Layer | May use | Must NOT use |
|-------|---------|--------------|
| Handler | `oapi.*`, API contract | `domain.*`, `model.*` |
| Service | API contract, `domain.*` | `oapi.*`, `model.*` |
| Repository | `domain.*`, `model.*` | `oapi.*`, API contract |

### Request Flow
```
oapi request → transform to service request → req.Validate() → req.CheckPermission() → 
domain model → repo (toModel) → go-jet query → repo (toDomain) → view → oapi response
```

### Error Rules
- Use `github.com/go-errors/errors` — NEVER `fmt.Errorf` or stdlib `errors`
- Six shared sentinels: `ErrNotFound`, `ErrValidation`, `ErrUnauthorized`, `ErrForbidden`, `ErrConflict`, `ErrInternal`
- Wrap with `%w`: `errors.Errorf("layer.service.op: %w", err)`
- Log once, at handler, with structured slog

## Common Mistakes to Avoid

1. **Raw SQL** — Use go-jet exclusively, no `db.Query()` or `db.Exec()`
2. **fmt.Errorf** — Use `errors.Errorf` from `github.com/go-errors/errors`
3. **stdlib errors** — Import `github.com/go-errors/errors` not `"errors"`
4. **Domain imports in handler** — Handler must not import `domain.*`
5. **Missing Validate()** — Every service call must start with `req.Validate()`
6. **Missing CheckPermission()** — After Validate(), call `req.CheckPermission()`
7. **Partial updates** — Repo receives full domain model, not loose fields
8. **Business logic in repo** — Repository only does data access, no rules

## File Naming Conventions

```
backend/internal/app/
├── services/<module>/
│   ├── api.go          # Request structs, Validate(), CheckPermission(), interface
│   ├── service.go      # Module's exported service implementation
│   └── service/
│       └── <operation>.go  # One function per file, named after operation
├── domain/
│   ├── errors.go       # Sentinel errors only
│   └── models.go       # Pure domain types + business methods
└── repository/<module>/
    ├── postgres.go     # Struct, New(), interface check
    ├── to_model.go     # toModel/toDomain mappings
    └── <operation>.go  # One function per file
```

## When Starting a New Feature

1. Read `.agents/rules/` to understand the layer you're implementing
2. Add new API method to `services/<module>/api.go` (request struct + view type)
3. Implement in `services/<module>/service/` — call Validate(), CheckPermission() first
4. Implement repo method in `repository/<module>/` — use go-jet only
5. Wire in `webserver.wireServices` — never per-service wiring

## Use /harness-review After Writing Code

The harness includes a review skill that checks for:
- Layer violations (wrong types at boundaries)
- Missing Validate()/CheckPermission() calls
- Incorrect error wrapping
- Raw SQL usage
- Type leaks across boundaries

Run `/harness-review` after implementing any new endpoint.