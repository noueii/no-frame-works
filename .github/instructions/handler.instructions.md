---
applyTo: "backend/**/handler/**"
---

# Handler Layer Conventions

Handlers are pure transformers between the OpenAPI contract and the service layer.

## Allowed types
- `oapi.*` (OpenAPI generated) and API contract types (`post.CreatePostRequest`, `post.PostView`)
- Must NOT use: `domain.*` models or `model.*` (go-jet/database) types

## Rules

1. **oapi types only** — Handler methods receive and return oapi-codegen generated request/response objects. No manual `json.Decode` from `http.Request` or writing to `http.ResponseWriter`.

2. **Transform, don't validate** — Map oapi fields to service request structs, call the module API, map results back to oapi response types. No business logic, no validation, no DB calls.

3. **Call the module API interface** — Handlers call the module's exported API interface (e.g. `post.PostAPI`), never a concrete service or repository. The handler struct field must be typed as the interface.

4. **Error mapping only** — Translate service errors to HTTP responses (e.g. `ErrNotFound` → 404). No `fmt.Errorf` wrapping, no error creation.

5. **No dependency creation** — All dependencies injected through the constructor. Never call `New()` inside a handler method.
