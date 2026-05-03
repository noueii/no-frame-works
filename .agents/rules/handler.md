# Handler Layer Review Rubric

You are reviewing handler code in a Go backend that uses oapi-codegen for strict typed HTTP handlers.

Handlers are **pure transformers** between the OpenAPI contract and the service layer. They receive oapi-generated request objects, map fields to service Op structs, call the module API, and map the result back to oapi response types. Nothing else.

## Allowed types

Handlers may only work with:
- **oapi-codegen generated types** (`oapi.*`) — for request/response
- **API Op types** from `services/api/` (e.g. `api.CreatePostOp`) — for calling services

Handlers must NOT import or use:
- Domain models (`domain.*`)
- Database/go-jet models (`model.*`)
- oapi types from api (requests are built inline)

## Rules

### 1. oapi types only

Handler methods must receive and return oapi-codegen generated request/response objects. No manual JSON decoding from `http.Request` or writing to `http.ResponseWriter`.

❌ Wrong:
```go
func (h *Handler) editUsername(w http.ResponseWriter, r *http.Request) {
    var body editUsernameBody
    json.NewDecoder(r.Body).Decode(&body)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

✅ Correct:
```go
func (h *Handler) PutUpdatePost(
    ctx context.Context,
    request oapi.PutUpdatePostRequestObject,
) (oapi.PutUpdatePostResponseObject, error) {
    return oapi.PutUpdatePost200JSONResponse(toOAPIPost(result)), nil
}
```

### 2. Transform to Op, don't validate

Handlers map oapi fields to service Op structs. No business logic, no validation, no DB calls, no permission checks.

❌ Wrong — handler validates:
```go
func (h *Handler) PostCreatePost(ctx context.Context, request oapi.PostCreatePostRequestObject) (oapi.PostCreatePostResponseObject, error) {
    if request.Body.Title == "" {
        return oapi.PostCreatePost400JSONResponse{Error: "title required"}, nil
    }
    // validation belongs in the service layer
}
```

❌ Wrong — handler calls repo directly:
```go
func (h *Handler) GetPost(ctx context.Context, request oapi.GetPostRequestObject) (oapi.GetPostResponseObject, error) {
    post, err := h.repo.FindByID(ctx, request.Id.String())
    // handlers must not call repositories directly
}
```

✅ Correct — handler builds Op with Request:
```go
func (h *Handler) PostCreatePost(ctx context.Context, request oapi.PostCreatePostRequestObject) (oapi.PostCreatePostResponseObject, error) {
    result, err := h.app.API().Posts.CreatePost(ctx, &api.CreatePostOp{
        Request: api.CreatePostRequest{
            Title:    request.Body.Title,
            Content:  request.Body.Content,
            AuthorID: a.UserID().String(),
        },
    })
    if err != nil {
        return oapi.PostCreatePost400JSONResponse{...}, nil
    }
    return oapi.PostCreatePost201JSONResponse(toOAPIPost(result)), nil
}
```

### 3. Access services through App.API()

Handlers access services via `h.app.API().Posts.CreatePost(...)`. The API struct holds typed service interfaces (not `any`).

❌ Wrong — concrete types:
```go
type Handler struct {
    postService *post.Service       // concrete type
}
```

✅ Correct — through App.API():
```go
h.app.API().Posts.CreatePost(ctx, &api.CreatePostOp{...})
```

### 4. Error mapping only

Handlers translate service errors to the appropriate HTTP response type. They do not create new errors or wrap errors.

❌ Wrong:
```go
if err != nil {
    return nil, fmt.Errorf("handler: failed to create post: %w", err)
}
```

✅ Correct:
```go
if err != nil {
    if errors.Is(err, apperrors.ErrNotFound) {
        return oapi.GetPost404JSONResponse{...}, nil
    }
    return oapi.GetPost400JSONResponse{...}, nil
}
```

### 5. No dependency creation

Handlers receive all dependencies through `*config.App`. They call `h.app.API().ServiceName.Method()` — services are wired once at startup, not per-request.

❌ Wrong:
```go
func (h *Handler) PostCreatePost(ctx context.Context, ...) {
    svc := post.New(app, repo)  // creating service per request
    result, err := svc.CreatePost(ctx, ...)
}
```

✅ Correct:
```go
// services wired in webserver.wireModules(), handler uses app
result, err := h.app.API().Posts.CreatePost(ctx, &api.CreatePostOp{...})
```

## Op Structure

Ops live in `services/api/` and have two parts:

```go
// services/api/post.go
type CreatePostOp struct {
    Request CreatePostRequest  // what caller provides
    Post    *domain.Post        // state populated by service
}
```

Handlers only set `Request`. Service populates state fields like `Post`.

## Output Format

Only flag violations where you are at least 80% confident. Skip rules that don't apply to the diff. When in doubt, don't flag it.

For each violation, provide:
- Rule name
- File path
- The problematic code or function
- Brief explanation of what's wrong