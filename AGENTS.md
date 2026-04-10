# Backend Architecture Conventions

This project follows a strict layered architecture. All agents (Claude Code, Copilot, CI reviewers) must follow these rules when writing or reviewing code.

Detailed rubrics with examples live in `.agents/rules/`.

## Type Isolation

Each layer has strict type boundaries:

| Layer | Uses | Must NOT use |
|-------|------|-------------|
| Handler | `oapi.*` + API contract (`Request`, `View`) | `domain.*`, `model.*` |
| Service | API contract + `domain.*` | `oapi.*`, `model.*` |
| Repository | `domain.*` + `model.*` (go-jet) | `oapi.*`, API contract |

## Request Flow

Every request follows this lifecycle:
1. Handler receives oapi request → transforms to service request struct
2. Service calls `req.Validate()` first → performs business logic with domain models
3. Repository receives complete domain model → converts to go-jet model via `toModel()` → executes query
4. Return path: repo returns domain model via `toDomain()` → service converts to view type → handler converts to oapi response

## Handler Layer
- Handlers are pure transformers between OpenAPI contract and service layer
- Only use oapi-codegen generated request/response objects — no manual JSON decode
- Call the module's API interface, never concrete services or repositories
- Error mapping only — translate service errors to HTTP responses
- Dependencies injected through constructor

## Service Layer
- Every service method calls `req.Validate()` first
- Services only accessible through the module's API interface
- Request structs own `Validate()` and `Permission()` methods in `api.go`
- One exported `Execute` function per file in service subfolders
- Domain model is source of truth — for updates: fetch existing → mutate fields → send complete model to repo
- Constructor injection for all dependencies

## Domain Layer
- No infrastructure imports (`database/sql`, `net/http`, external SDKs)
- Methods on domain types must be pure business logic only
- Sentinel errors in `domain/errors.go` — not inline, not in module root
- Types owned by their module — no shared domain packages
- No persistence or presentation concerns (no `ToJSON`, `ToDBModel`)

## Repository Layer
- No raw SQL — use go-jet query builder exclusively
- Updates use `MODEL()` with `MutableColumns` — exceptions must be commented
- Mapping functions (`toModel`/`toDomain`) live in the repository
- No business logic — only data access
- Model-in, model-out — simple CRUD with complete domain models, no field-specific methods
- One function per file in subfolders

## Common Patterns
- Use `errors.New()` for static errors, `fmt.Errorf()` only for wrapping with `%w`
- Always wrap errors with `%w`, not `%v` or `%s`
- Return sentinel errors directly — don't wrap domain errors
