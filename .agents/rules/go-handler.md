---
applies-to: ["*.go", "internal/webserver/handler/**/*.go"]
---

# Go Handler Layer Rules

## Structure
- One file per entity: `get_post.go`, `create_post.go`, etc.
- `handler.go`: thin wrapper with `h.app.API().Service.Method()`

## Responsibilities
- **ONLY**: unwrap oapi request → call service → map errors to HTTP status
- **NO** business logic
- **NO** direct repo access
- **NO** domain method calls beyond read-only getters

## Type Isolation
- Use `oapi.*` types from generated code
- Use service API contract types (Request/View from `services/api/`)
- **MUST NOT** import `domain.*` or `model.*`

## Error Mapping
```go
switch {
case errors.Is(err, apperrors.ErrNotFound):
    w.WriteHeader(http.StatusNotFound)
case errors.Is(err, apperrors.ErrValidation):
    w.WriteHeader(http.StatusBadRequest)
// ...
}
```

## Pattern
```go
func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
    req := oapi.GetPostFromRequest(r)
    op := &api.GetPostOp{Request: req}
    
    view, err := h.app.API().Posts.GetPost(r.Context(), op)
    if err != nil {
        // map error to status
        return
    }
    
    // write response
}
```

## Dependencies
- Handler holds `*config.App` only
- All service calls via `h.app.API().Service.Method()`