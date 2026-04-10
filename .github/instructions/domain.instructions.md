---
applyTo: "backend/**/domain/**"
---

# Domain Layer Conventions

The domain layer defines data structures and business rules. It has zero knowledge of infrastructure.

## Rules

1. **No infrastructure imports** — No `database/sql`, `net/http`, external SDKs, or framework code. Only standard library types (`time`, `errors`, `fmt`, `strings`).

2. **Pure business logic only** — Methods on domain types must be computations, state checks, or transformations. No I/O, no database calls, no HTTP calls.

3. **Sentinel errors in domain/errors.go** — Module-specific errors defined as `var` sentinel errors in the module's `domain/errors.go`. Not inline `fmt.Errorf`, not in the module root, not in service files.

4. **Types owned by their module** — No shared domain packages. Modules must not import another module's `domain/` package. Reference other entities by ID, not by importing their type.

5. **No persistence or presentation concerns** — No `ToJSON()`, `ToDBModel()`, or mapping functions. Mapping belongs in the repository (`toModel`/`toDomain`) or handler layer.
