# Architecture Overview

This project uses a strict four-layer architecture with hard type boundaries. Every request flows through the same shape: **handler → service → domain → repository**, and every layer has a single responsibility. Cross-layer leakage is the most common cause of bugs, so the rules are enforced rather than suggested.

These docs are for engineers joining the project. Agent-facing rubrics live in [`.agents/rules/`](../../.agents/rules/) — those are the terse versions of the same rules.

## The Layers

```
┌─────────────────────────────────────────────────┐
│  HANDLER   internal/webserver/handler/          │
│    oapi request → service request → oapi response
└─────────────────────────────────────────────────┘
                       │
┌─────────────────────────────────────────────────┐
│  SERVICE   internal/modules/<mod>/service/      │
│    Validate → CheckPermission → fetch/mutate/save
└─────────────────────────────────────────────────┘
                       │
┌─────────────────────────────────────────────────┐
│  DOMAIN    internal/modules/<mod>/domain/       │
│    pure types + business rules, no I/O          │
└─────────────────────────────────────────────────┘
                       │
┌─────────────────────────────────────────────────┐
│  REPOSITORY   repository/<mod>/                 │
│    go-jet queries, toModel/toDomain mapping     │
└─────────────────────────────────────────────────┘
```

## Type Isolation

Each layer may only touch the types listed in its column. This is the most important rule in the project — violating it creates coupling you cannot undo cheaply.

| Layer      | Uses                                      | Must NOT use              |
|------------|-------------------------------------------|---------------------------|
| Handler    | `oapi.*` + module API contract            | `domain.*`, `model.*`     |
| Service    | Module API contract + `domain.*`          | `oapi.*`, `model.*`       |
| Repository | `domain.*` + `model.*` (go-jet generated) | `oapi.*`, API contract    |

The **API contract** is the exported surface of a module (the `api.go` file at the module root): the `API` interface, request structs, and the `View` return type. See [05-api-contract.md](05-api-contract.md).

## Request Lifecycle

Every request follows the same script. You should be able to read any handler and predict what happens next.

1. **Handler** receives an oapi-generated request object. It extracts the actor from context, builds a service-layer request struct, and calls the module's `API` interface.
2. **Service** calls `req.Validate()` first, then `req.CheckPermission(ctx[, model])`. For updates it fetches the existing domain model, mutates fields in memory, and sends the **complete** model to the repository.
3. **Repository** takes the domain model, converts it to a go-jet model via `toModel()`, runs the query, converts the result back via `toDomain()`, and returns a domain model.
4. **Return path**: repo returns domain → service converts to `*View` pointer → handler converts to an oapi response. On errors, the service returns `nil, err`; the handler maps sentinel errors to HTTP status codes via `errors.Is`.

There is no `fmt.Errorf`, no `errors.New`, no stdlib `errors` anywhere. Use `github.com/go-errors/errors` consistently — `errors.Errorf("...: %w", err)` for wrapping, `errors.Is` for comparison.

## Canonical Example

Throughout these docs, the `post` module is used as the reference example. It has the full CRUD shape, real authorization rules, and uses every pattern you'll need.

- Handler: `backend/internal/webserver/handler/post_create_post.go`
- Module contract: `backend/internal/modules/post/api.go`
- Service: `backend/internal/modules/post/service/`
- Domain: `backend/internal/modules/post/domain/`
- Repository: `backend/repository/post/`

## Reading Order

1. [00-overview.md](00-overview.md) — you are here
2. [01-handler.md](01-handler.md)
3. [02-service.md](02-service.md)
4. [03-domain.md](03-domain.md)
5. [04-repository.md](04-repository.md)
6. [05-api-contract.md](05-api-contract.md) — the module boundary that ties it all together
7. [06-actor.md](06-actor.md) — identity and how it propagates
8. [07-sentinel-errors.md](07-sentinel-errors.md) — how errors cross layer boundaries
9. [08-transactions.md](08-transactions.md) — atomic multi-repo operations
