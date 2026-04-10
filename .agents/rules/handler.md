# Handler Layer Review Rubric

You are reviewing handler code in a Go backend that uses oapi-codegen for strict typed HTTP handlers.

Handlers are **pure transformers** between the OpenAPI contract and the service layer. They receive oapi-generated request objects, map fields to service request structs, call the module API, and map the result back to oapi response types. Nothing else.

## Allowed types

Handlers may only work with:
- **oapi-codegen generated types** (`oapi.*`) — for request/response
- **Module API contract types** (e.g. `post.CreatePostRequest`, `post.PostView`) — for calling services

Handlers must NOT import or use:
- Domain models (`domain.*`)
- Database/go-jet models (`model.*`)

## Rules

### 1. oapi types only

Handler methods must receive and return oapi-codegen generated request/response objects. No manual JSON decoding from `http.Request` or writing to `http.ResponseWriter`.

❌ Wrong:
```go
func (h *Handler) editUsername(w http.ResponseWriter, r *http.Request) {
    var body editUsernameBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    // ...
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
    // ...
    return oapi.PutUpdatePost200JSONResponse(toOAPIPost(result)), nil
}
```

### 2. Transform, don't validate

Handlers map oapi fields to service request structs and map service results back to oapi response types. No business logic, no validation, no DB calls, no permission checks.

❌ Wrong:
```go
func (h *Handler) PostCreatePost(ctx context.Context, request oapi.PostCreatePostRequestObject) (oapi.PostCreatePostResponseObject, error) {
    if request.Body.Title == "" {
        return oapi.PostCreatePost400JSONResponse{Error: "title required"}, nil
    }
    // validation belongs in the service layer
}
```

❌ Wrong:
```go
func (h *Handler) GetPost(ctx context.Context, request oapi.GetPostRequestObject) (oapi.GetPostResponseObject, error) {
    post, err := h.repo.FindByID(ctx, request.Id.String())
    // handlers must not call repositories directly
}
```

✅ Correct:
```go
func (h *Handler) PostCreatePost(ctx context.Context, request oapi.PostCreatePostRequestObject) (oapi.PostCreatePostResponseObject, error) {
    result, err := h.postAPI.CreatePost(ctx, post.CreatePostRequest{
        Title:    request.Body.Title,
        Content:  request.Body.Content,
        AuthorID: a.UserID().String(),
    })
    if err != nil {
        return oapi.PostCreatePost400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse{Error: err.Error()}}, nil
    }
    return oapi.PostCreatePost201JSONResponse(toOAPIPost(result)), nil
}
```

### 3. Call the module API interface

Handlers call the module's exported API interface (e.g. `post.PostAPI`), never a concrete service or repository directly. The handler struct field must be typed as the interface.

❌ Wrong:
```go
type Handler struct {
    postService *service.Service       // concrete type
    postRepo    *postrepo.PostgresRepo  // direct repo access
}
```

✅ Correct:
```go
type Handler struct {
    postAPI post.PostAPI  // interface
}
```

### 4. Error mapping only

Handlers translate service errors to the appropriate HTTP response type. They do not create new errors or wrap errors with `fmt.Errorf`.

❌ Wrong:
```go
if err != nil {
    return nil, fmt.Errorf("handler: failed to create post: %w", err)
}
```

✅ Correct:
```go
if err != nil {
    if errors.Is(err, post.ErrPostNotFound) {
        return oapi.GetPost404JSONResponse{Error: "post not found"}, nil
    }
    return oapi.GetPost400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse{Error: err.Error()}}, nil
}
```

### 5. No dependency creation

Handlers receive all dependencies through the constructor. They never call `New()` or instantiate services/repos internally.

❌ Wrong:
```go
func (h *Handler) PostCreatePost(ctx context.Context, request oapi.PostCreatePostRequestObject) (oapi.PostCreatePostResponseObject, error) {
    repo := postrepo.New(h.app.DB())
    svc := postservice.New(repo)
    result, err := svc.CreatePost(ctx, ...)
}
```

✅ Correct:
```go
// Dependencies injected at construction time
func NewHandler(app *config.App) *Handler {
    repo := postrepo.New(app.DB())
    svc := postservice.New(repo)
    return &Handler{
        postAPI: svc,
    }
}
```

## Output Format

Only flag violations where you are at least 80% confident. Skip rules that don't apply to the diff. When in doubt, don't flag it.

For each violation, provide:
- Rule name
- File path
- The problematic code or function
- Brief explanation of what's wrong
