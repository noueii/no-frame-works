# Handler Layer

Handlers live in `backend/internal/webserver/handler/`. They are **pure transformers** between the OpenAPI contract and the service layer. If a handler contains a conditional that isn't about HTTP status mapping, it probably belongs in the service.

## Responsibilities

A handler does exactly four things, in order:

1. **Extract the actor** from context.
2. **Transform** the oapi request object into a service-layer request struct.
3. **Call the module's `API` interface** — never the concrete service, never a repository.
4. **Map the result** back to an oapi response — domain model errors become HTTP status codes.

That's the whole job. If you find yourself doing anything else in a handler — validating fields, loading related data, composing multiple service calls into a business operation — stop and move it into the service.

## What Handlers Must NOT Do

- No `domain.*` or `model.*` imports. Handlers don't know about domain types or database models.
- No manual JSON decoding. oapi-codegen generates request objects; use them.
- No permission checks. Authorization lives on the request struct (`CheckPermission`) and is executed by the service.
- No validation. Validation lives on the request struct (`Validate`) and is executed by the service.
- No direct repository access. The only collaborator is the module's `API` interface.

## Anatomy of a Handler

From `backend/internal/webserver/handler/post_create_post.go`:

```go
func (h *Handler) PostCreatePost(
    ctx context.Context,
    request oapi.PostCreatePostRequestObject,
) (oapi.PostCreatePostResponseObject, error) {
    a := actor.From(ctx)
    if a == nil {
        return oapi.PostCreatePost401JSONResponse{
            Error: "unauthorized",
        }, nil
    }

    result, err := h.postAPI.CreatePost(ctx, post.CreatePostRequest{
        Title:    request.Body.Title,
        Content:  request.Body.Content,
        AuthorID: a.UserID().String(),
    })
    if err != nil {
        if errors.Is(err, post.ErrUnauthorized) {
            return oapi.PostCreatePost401JSONResponse{
                Error: "unauthorized",
            }, nil
        }
        return oapi.PostCreatePost400JSONResponse{
            ErrorJSONResponse: oapi.ErrorJSONResponse{Error: err.Error()},
        }, nil
    }

    return oapi.PostCreatePost201JSONResponse(toOAPIPost(*result)), nil
}
```

Read it top-to-bottom: actor check → build service request → call `h.postAPI` → map error → return oapi response. Nothing else.

## Actor Extraction

The first thing almost every handler does is extract the actor:

```go
a := actor.From(ctx)
if a == nil {
    return oapi.<Endpoint>401JSONResponse{Error: "unauthorized"}, nil
}
```

`actor.From(ctx)` returns `nil` when no actor is present (e.g. an unauthenticated request slipped through middleware). The actor middleware in `internal/webserver/middleware/auth_middleware.go` attaches an actor to every authenticated request. See [06-actor.md](06-actor.md) for the full story.

Handlers pass the actor's ID into the service request struct (`AuthorID: a.UserID().String()`) when the service needs to record "who did this." They do **not** pass the full actor — the service re-extracts it from context inside `CheckPermission` because authorization is the request struct's job, not the handler's.

## Error Mapping

Handlers translate sentinel errors from the service into HTTP status codes using `errors.Is`:

```go
if errors.Is(err, post.ErrPostNotFound) {
    return oapi.GetPost404JSONResponse{Error: "post not found"}, nil
}
if errors.Is(err, post.ErrForbidden) {
    return oapi.GetPost403JSONResponse{Error: "forbidden"}, nil
}
return oapi.GetPost500JSONResponse{Error: "internal error"}, nil
```

This is the **only** layer that decides HTTP status codes. Services return domain-shaped errors; handlers map them. If you find yourself wanting a new status code, that's a hint you may need a new sentinel error — see [07-sentinel-errors.md](07-sentinel-errors.md).

Note the import: handlers import the **module package** (`post.ErrPostNotFound`), not the domain package. The module re-exports the domain sentinels so handlers don't have to cross the type isolation boundary.

## Dependency Injection

Handlers receive their dependencies through a constructor — typically a single `Handler` struct wired up at startup:

```go
type Handler struct {
    postAPI post.API
    userAPI user.API
    // ...
}
```

The fields are **interface types** (`post.API`), not concrete services. This is what makes handlers testable in isolation and what prevents them from reaching past the module boundary.

## Shape of a New Handler

When adding a new endpoint, copy an existing handler in the same module and change four things: the oapi types, the service request struct, the service method, and the error→status mapping. Don't invent new shapes.
