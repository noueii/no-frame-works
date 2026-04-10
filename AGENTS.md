# Backend Architecture Conventions

This project follows a strict layered architecture. All agents (Claude Code, Copilot, CI reviewers) must follow these rules when writing or reviewing code.

Detailed rubrics with examples live in `.agents/rules/`.

## Communication Style

Follow the caveman communication style defined in `.agents/caveman.md` — terse, no filler, all technical substance preserved.

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
2. Service calls `req.Validate()` first → `req.CheckPermission(...)` second → performs business logic with domain models
3. Repository receives complete domain model → converts to go-jet model via `toModel()` → executes query
4. Return path: repo returns domain model via `toDomain()` → service converts to `*View` pointer → handler converts to oapi response

## Handler Layer
- Handlers are pure transformers between OpenAPI contract and service layer
- Only use oapi-codegen generated request/response objects — no manual JSON decode
- Call the module's API interface, never concrete services or repositories
- Error mapping only — translate service errors to HTTP responses
- Dependencies injected through constructor

## Service Layer
- Every service function calls `req.Validate()` first, then `req.CheckPermission(...)` — no separate permission middleware
- Services only accessible through the module's API interface
- Request structs own `Validate()` and `CheckPermission()` methods in `api.go`
- One exported function per file in service subfolders — named after the operation (not `Execute`)
- Domain model is source of truth — for updates: fetch existing → mutate fields → send complete model to repo
- Return `*View` pointers — on error, return `nil, err` (never empty structs)
- Constructor injection for all dependencies

## Domain Layer
- No infrastructure imports (`database/sql`, `net/http`, external SDKs)
- Methods on domain types must be pure business logic only
- Authorization rules live on domain models as pure methods (e.g. `CanModify(actor) bool`)
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
- Never use `fmt.Errorf` or stdlib `"errors"` — use `"github.com/go-errors/errors"` everywhere
- Use `errors.Errorf()` for wrapping (`%w`) and for sentinel error declarations
- Always wrap errors with `%w`, not `%v` or `%s`
- Return sentinel errors directly — don't wrap domain errors
