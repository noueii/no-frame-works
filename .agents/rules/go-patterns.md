---
applies-to: ["*.go", "internal/**/*.go", "cmd/**/*.go"]
---

# Go Project Patterns

## Error Handling
- **Always use**: `github.com/go-errors/errors`
- **Never use**: `fmt.Errorf` or stdlib `"errors"`
- Wrap pattern: `errors.Errorf("layer.service.op: %w", err)`
- Log once at handler, never swallow silently

## Type Isolation Layers
| Layer | Uses | Must NOT use |
|-------|------|--------------|
| Handler | `oapi.*`, service API contract | `domain.*`, `model.*` |
| Service | API contract, `domain.*` | `oapi.*`, `model.*` |
| Repository | `domain.*`, `model.*` | `oapi.*`, service API |

## Request Flow
1. Handler unwraps oapi request
2. Service calls `req.Validate()` then `req.CheckPermission()`
3. Service calls repository
4. Repo returns domain model via `toDomain()`
5. Service converts to `*View` pointer
6. Handler converts to oapi response

## Configuration
- `config.App`: holds `*API{Posts api.PostAPI, Users api.UserAPI}`
- Typed APIs - no type assertions needed in handlers
- Constructor injection for all dependencies

## File Naming
- Operations: `create_post.go`, `get_post.go` (underscore)
- Domain: `post.go`, `user.go`
- Service API: `api.go` (for Request/Op types)

## Imports
- No circular dependencies
- Neutral `services/api/` package breaks cycles (Request/Op types)
- Cross-service calls via `s.app.API().Other.X()`