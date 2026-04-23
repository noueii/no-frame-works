# Architecture Documentation

Engineer-facing documentation for the `no-frame-works` backend architecture.

## Read This First

👉 **[application-pattern.md](application-pattern.md)** — the canonical description of how this codebase is structured.

This is the **current architecture pattern**. Every service, handler, and module in the codebase follows it. New features must follow it. PR reviews should enforce it. If there's a conflict between `application-pattern.md` and anything in the per-layer docs below, `application-pattern.md` wins.

The short version (60 seconds):

- One container — `*config.App` — carries every cross-cutting dependency and a `*config.API` struct holding each module's service interface.
- Services have a uniform shape: `type Service struct { app *config.App; repo ModuleRepository }`. Two fields, always.
- Same-module data access via `s.repo`. Cross-module access via `s.app.API().Other.X(...)`. There is **no way** to reach another module's repository — the `App` doesn't expose one. This is compile-time enforced.
- Handlers hold `*config.App` and call `h.app.API().Module.X(...)`. Nothing else.
- Wiring is centralized in `webserver.wireModules` — one function, one place, one flow.
- Testing uses hand-rolled function-field stubs plus a `testapp.New()` helper. Per-test setup is 3–5 lines and does not scale with the number of modules.

Read [application-pattern.md](application-pattern.md) for the full story: rules, reasoning, code examples, testing patterns, and a checklist for adding new modules.

## Per-layer deep dives (reference)

These files were written earlier and document specific layers of the architecture in depth. They are still mostly correct, but where they describe constructor-injection-style service construction, they are **superseded by `application-pattern.md`**. Read them for detail on the layer you're working in, not as the primary architecture reference.

| # | Doc | Topic |
|---|-----|-------|
| 00 | [Overview](00-overview.md) | Layered architecture, type isolation, request lifecycle |
| 01 | [Handler Layer](01-handler.md) | oapi boundary, actor extraction, error → HTTP mapping |
| 02 | [Service Layer](02-service.md) | Validation, permission, orchestration (partially superseded — see `application-pattern.md` §6) |
| 03 | [Domain Layer](03-domain.md) | Pure types, authorization methods, no infrastructure |
| 04 | [Repository Layer](04-repository.md) | go-jet, `toModel`/`toDomain`, domain-in/domain-out |
| 05 | [Module API Contract](05-api-contract.md) | `api.go`: the `API` interface, request structs, `View` |
| 06 | [Actor](06-actor.md) | Identity propagation from middleware to domain methods |
| 07 | [Sentinel Errors](07-sentinel-errors.md) | Declaring, returning, and matching errors across layers |
| 08 | [Transactions](08-transactions.md) | `TxManager`, `GetExecutor`, atomic multi-repo operations (not yet implemented — noted as a future addition in `application-pattern.md` §16) |

## Pain-point analysis (reference)

[issues/](issues/) contains a layer-by-layer comparison of how three existing codebases (`intranet-cms`, `workflows`, `id-services v3`) handle the same concerns, with good-and-bad callouts and per-repo improvements. These are the inputs that drove the design of `application-pattern.md` — read them if you want to understand why specific rules exist in terms of concrete pain they prevent.

- [issues/00-overview.md](issues/00-overview.md) — cross-codebase comparison overview
- [issues/01-handler-layer.md](issues/01-handler-layer.md) — handler layer across repos
- [issues/02-application-layer.md](issues/02-application-layer.md) — service/application layer across repos
- [issues/03-repository-layer.md](issues/03-repository-layer.md) — repository layer across repos
- [issues/04-permission-duplication.md](issues/04-permission-duplication.md) — the permission-check duplication problem

## Canonical example

All code examples throughout these docs use the **`post` module** in the current codebase. If you want to see the whole pattern working end-to-end:

- **Handlers**: `backend/internal/webserver/handler/post_*.go`
- **Module contract**: `backend/internal/modules/post/api.go`, `errors.go`, `repository.go`
- **Domain**: `backend/internal/modules/post/domain/`
- **Service**: `backend/internal/modules/post/service/service.go`
- **Repository**: `backend/repository/post/`
- **Wiring**: `backend/internal/webserver/webserver.go` (the `wireModules` function)
- **App container**: `backend/config/app.go`, `backend/config/api.go`

A cross-module example lives between `post.Service.CreatePost` and `user.Service.IncrementPostCount` — post creates a post and then calls `s.app.API().User.IncrementPostCount` to keep the author's denormalized post count in sync.
