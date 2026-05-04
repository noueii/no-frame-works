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

## Errors
- Use `github.com/go-errors/errors` everywhere. Never `fmt.Errorf`, never stdlib `"errors"`. `errors.Errorf` and `errors.Is/As` from go-errors are `%w`-compatible with stdlib and additionally capture stack traces.
- **Six shared sentinels** in `internal/app/apperrors`: `ErrNotFound`, `ErrValidation`, `ErrUnauthorized`, `ErrForbidden`, `ErrConflict`, `ErrInternal`. These map to HTTP status codes. No per-service sentinels (no `post.ErrPostNotFound`, etc.).
- **User-facing errors use `*apperrors.Coded`** constructed via `apperrors.Validation(code, message, params)`, `apperrors.NotFound(...)`, `apperrors.Conflict(...)`, `apperrors.Forbidden(...)`, `apperrors.Unauthorized(...)`. Each helper wires the sentinel kind automatically.
- **Error codes are typed constants** in `apperrors/codes.go`, shared with the frontend as translation keys. Naming: `<service>.<specific_reason>` (e.g., `CodePostTitleRequired = "post.title_required"`).
- **Wrap at every layer** with `errors.Errorf` using the `layer.service.operation: context: %w` convention. Examples: `"repo.post.FindByID id=abc: %w"`, `"service.post.CreatePost: load existing: %w"`, `"domain.post.Validate: title required: %w"`. The layer prefix makes error chains self-labeling in logs.
- **Handlers match sentinels** with `errors.Is` to pick HTTP status codes, and extract `Code`/`Params` via `apperrors.Message(err, fallback)`, `apperrors.CodeOf(err)`, `apperrors.ParamsOf(err)` for the response body.
- **Log errors once, at the handler**, with structured slog attributes (request ID, actor, error chain via `slog.Any("error", err)`, error code via `slog.String("error_code", apperrors.CodeOf(err))`).
- **Never swallow errors silently.** Either return them up the stack or log them explicitly.
- See `docs/architecture/application-pattern.md` §16 for the full pattern with code examples.

## Type Isolation (layer rules)

- **Handler**: `oapi.*` + service API contract (request/view types from `services/<name>/api.go`). Must NOT import `domain.*` or storage `model.*`.
- **Service**: service API contract + `domain.*` + `apperrors`. Must NOT import `oapi.*` or storage `model.*`.
- **Repository**: `domain.*` + `model.*` (go-jet) + `apperrors` for sentinels. Must NOT import `oapi.*` or service API contract.
- **Domain**: pure types and pure methods. No `database/sql`, no `net/http`, no external SDKs, no context-dependent I/O. Cross-service domain imports are allowed when the target is a stable leaf (e.g. `services/post/domain/Y` importing `services/user/domain/X`), but in this codebase the domain is consolidated under `internal/app/domain/` so cross-service domain imports don't arise.

## Service Layer
- Services have the shape `type Service struct { app *config.App; repo XRepository }`. Two fields, always. Constructor takes both: `New(app, repo)`.
- Every service method calls `req.Validate()` first; request structs own their own validation in `api.go`.
- Cross-service operations use `s.app.API().Other.X(...)`. Direct cross-service repo access is compile-impossible (the App doesn't expose `Repos()`).
- Services return their own module's domain types only. No composed types spanning services (composition happens at the handler).
- Wiring lives only in `webserver.wireServices`. No per-service wiring, no init functions.

## Repository Layer
- Use go-jet exclusively. No raw SQL.
- Updates use `MODEL(MutableColumns)`. Exceptions must be commented.
- Mapping functions (`toModel`/`toDomain`) are private to the repository package.
- Domain-in, domain-out. No partial-update methods. The service hands the repo a complete domain value.
- One function per file in per-entity sub-packages.

## Handler Layer
- Thin. Hold `*config.App` only. Every handler reads `h.app.API().Service.X(...)` per call.
- Handler methods: unwrap oapi, call the service, map errors to HTTP status via `errors.Is` on `apperrors.*` sentinels.
- No business logic in handlers. No repo access. No domain method calls beyond read-only getters.
