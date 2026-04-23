# Actor (Identity Propagation)

The **Actor** is the project's answer to "who is doing this?" It's a small interface defined in `backend/internal/core/actor/actor.go` that gets attached to every authenticated request and threads all the way down to domain-level authorization methods.

Actors live in `internal/core/`, not in any module. They are a cross-cutting primitive — every module needs to know "who's acting" but no module owns the concept.

## The Interface

```go
type Actor interface {
    IsSystem() bool
    UserID() uuid.UUID
}
```

That's it. Two methods. Everything else is implementation.

Two concrete types implement it:

```go
type UserActor struct {
    ID   uuid.UUID
    Role Role
}

func (a UserActor) IsSystem() bool    { return false }
func (a UserActor) UserID() uuid.UUID { return a.ID }
func (a UserActor) HasRole(r Role) bool { return a.Role == r }

type SystemActor struct {
    Service string
}

func (a SystemActor) IsSystem() bool    { return true }
func (a SystemActor) UserID() uuid.UUID { return uuid.Nil }
```

`UserActor` is a human making a request through the HTTP layer. `SystemActor` is an internal process — a background job, a webhook handler, a migration running as "itself." Both satisfy `Actor`, so code that only cares about "someone is acting" can take the interface; code that needs user-specific behavior can type-assert to `UserActor`.

`Role` is a string type defined in the same package (`RoleAdmin`, `RoleMember`). It lives on `UserActor` because roles are a human concept — `SystemActor` doesn't have a role, it has unconditional system privileges.

## Origination: the Auth Middleware

Actors enter the system exactly once, in `backend/internal/webserver/middleware/auth_middleware.go`:

```go
func NewActorMiddleware(idClient identity.Client) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if strings.HasPrefix(r.URL.Path, "/api/v1/auth/") {
                next.ServeHTTP(w, r)
                return
            }

            sessionCookie, err := r.Cookie("ory_kratos_session")
            if err != nil || sessionCookie.Value == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            detail, err := idClient.GetSession(r.Context(), sessionCookie.Value)
            if err != nil {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            userID, err := uuid.Parse(detail.IdentityID)
            if err != nil {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            userActor := actor.UserActor{ID: userID, Role: actor.RoleMember}
            ctx := actor.WithActor(r.Context(), userActor)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

The middleware reads the session cookie, asks the identity provider (Ory Kratos) who the user is, constructs a `UserActor`, and attaches it to the request context with `actor.WithActor(ctx, userActor)`. From that point on, every handler and service downstream can call `actor.From(ctx)` and get the actor back.

Public auth endpoints (`/api/v1/auth/*`) are excluded — you don't have an identity yet when you're logging in.

## Extraction: `actor.From(ctx)`

```go
func WithActor(ctx context.Context, a Actor) context.Context {
    return context.WithValue(ctx, contextKey{}, a)
}

func From(ctx context.Context) Actor {
    a, _ := ctx.Value(contextKey{}).(Actor)
    return a
}
```

The contract is: **`From` returns `nil` when no actor is present.** This is not an error; it's a cue to the caller that the request is either pre-authentication or malformed. Every handler checks for `nil` and returns 401 if so; every `CheckPermission` method does the same and returns `ErrUnauthorized`.

`contextKey{}` is an unexported empty struct used as the map key. This prevents collisions with any other `context.WithValue` in the system — two packages can't accidentally share a key.

## Propagation Through the Layers

The actor is **always** passed via context, never as an argument to service or repository calls. The flow looks like:

```
Middleware    →  actor.WithActor(ctx, userActor)
Handler       →  a := actor.From(ctx); h.postAPI.CreatePost(ctx, req)
Service       →  (called by handler, passes ctx through)
Request struct→  func (r X) CheckPermission(ctx) { a := actor.From(ctx); ... }
Domain method →  func (p Post) CanModify(a actor.Actor) bool { ... }
```

Notice where the actor **stops** being in context: at the boundary of a domain method. Domain methods take the actor as an **argument**, not from ctx. This is a deliberate rule.

## Why Domain Methods Take Actor as an Argument

Three reasons:

1. **Domain code never imports `context`.** The domain layer is pure business logic — it doesn't know about request lifecycles, cancellation, or deadlines. Taking `ctx` would leak infrastructure into the domain.
2. **Tests don't need a fake context.** `post.CanModify(actor.UserActor{...})` is trivially testable. `post.CanModify(ctxWithActor)` would force every test to build a context.
3. **Authorization is explicit at the call site.** When you read `post.CanModify(a)`, you know exactly what identity is being checked. When you read `post.CanModify(ctx)`, you don't — you have to trust that the context has the right actor.

The request struct's `CheckPermission` method is the translator: it pulls the actor out of context once, and hands it to the domain method as an argument.

```go
func (r UpdatePostRequest) CheckPermission(ctx context.Context, post *domain.Post) error {
    a := actor.From(ctx)
    if a == nil {
        return ErrUnauthorized
    }
    if !post.CanModify(a) {  // ← pure domain call, actor is explicit
        return ErrForbidden
    }
    return nil
}
```

This is the only layer that does `actor.From(ctx)`. Everything below it operates on a concrete `actor.Actor` value.

## System Actor

`SystemActor` is used for internal code paths that don't have a human user — background jobs, migrations, webhook consumers. The typical pattern is to construct one at the entry point of the job and attach it to a fresh context:

```go
ctx := actor.WithActor(context.Background(), actor.SystemActor{Service: "digest-worker"})
```

Domain methods short-circuit system actors to allow-all in most cases:

```go
func (p Post) CanModify(a actor.Actor) bool {
    if a.IsSystem() {
        return true
    }
    // ... user-specific rules
}
```

This reflects the policy: system code is trusted; user code is not. If a particular domain rule should **not** allow system actors (rare), the predicate can check `if a.IsSystem() { return false }` — but you almost never want that.
