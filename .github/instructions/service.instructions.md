---
applyTo: "backend/**/service/**,backend/**/api.go,backend/**/middleware/**"
---

# Service Layer Conventions

Services contain business logic and are accessed exclusively through the module's API interface.

## Allowed types
- API contract types (request structs + view types from `api.go`) and `domain.*` models
- Must NOT use: `oapi.*` (OpenAPI generated) or `model.*` (go-jet/database) types

## Rules

1. **Validate first** — Every service method or Execute function must call `req.Validate()` as its first meaningful operation.

2. **Accessed through API interface only** — Services must only be called through the module's exported API interface. No external code should import a service package directly.

3. **Request structs own Validate() and Permission()** — Every request struct implements both methods, defined in the module's `api.go`.

4. **One function per file** — Each file in a service subfolder contains exactly one exported `Execute` function. Only `service/service.go` can have multiple methods.

5. **Domain types internally, view types externally** — Services work with domain models internally but accept request structs as input and return view types as output.

6. **Constructor injection** — Dependencies injected through constructors, never created internally.

7. **Domain model is source of truth** — For updates: fetch the existing model from the repo, mutate fields directly in memory, send the complete model to the repo. Never send loose fields or partial objects. Cross-module side effects are orchestrated by the service, not the model.
