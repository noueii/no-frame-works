# Application Architecture: The God-App Pattern

This document is the **canonical description of how this codebase is structured**. It supersedes the per-layer docs (`01-handler.md`, `02-service.md`, etc.) where they conflict. New features must follow this pattern. Review feedback on PRs should enforce it.

If you're reading this as a new engineer: read the whole thing once end-to-end, then keep it handy as a reference. If you're reviewing a PR: changes that depart from this pattern should have an explicit reason in the PR description.

## Contents

1. [The pattern at a glance](#1-the-pattern-at-a-glance)
2. [Why this shape (the design drivers)](#2-why-this-shape)
3. [The `App` container](#3-the-app-container)
4. [The `API` container](#4-the-api-container)
5. [Service layout](#5-service-layout)
6. [The `Service` struct](#6-the-service-struct)
7. [Cross-service calls](#7-cross-service-calls)
8. [Request struct methods](#8-request-struct-methods)
9. [Handlers](#9-handlers)
10. [Repositories](#10-repositories)
11. [Wiring](#11-wiring)
12. [Import graph and why it's acyclic](#12-import-graph-and-why-its-acyclic)
13. [Testing](#13-testing)
14. [The hard rules](#14-the-hard-rules)
15. [Adding a new service — a checklist](#15-adding-a-new-service)
16. [Known gaps and future work](#16-known-gaps-and-future-work)

---

## 1. The pattern at a glance

- **One container** — `*config.App` — carries every cross-cutting dependency the app needs: the database, the logger, the identity provider, and a `*config.API` struct that holds the service interfaces for every service.
- **Services** hold `*config.App` **and** their own repository as separate struct fields. Same-service data access goes through `s.repo`; cross-service operations go through `s.app.API().Other.X(...)`.
- **Repositories are NOT on the `App`.** This is a hard rule, enforced at compile time. There is no `app.Repos()` accessor. The only way for one service to touch another service's state is through the other service's public API, which runs inside the other service's implementation (where invariants, validation, and business rules live).
- **Handlers are thin.** They unwrap the oapi request object, call `h.app.API().Post.CreatePost(...)`, and map errors to HTTP status codes via `errors.Is` on sentinel errors.
- **Wiring is centralized.** `webserver.wireServices` constructs every repo and service in one place and registers the service APIs on the `App` before any handler is built.
- **Tests use stubs + a `testapp.New()` helper.** Function-field stubs are written once per API interface (hand-rolled or code-generated) and reused across every test. Per-test setup is 3–5 lines and does not scale with the number of services in the project.

## 2. Why this shape

Every rule in this document is a response to a specific problem. Before reading the rest, understand the forces:

**Driver 1 — Request-scoped methods need cheap access to many dependencies.**

Request types in the service root (e.g. `post.CreatePostRequest`) grow methods over time: `Validate()`, `CheckPermission()`, `CheckExisting()`, `NotifyMentions()`, `Index()`, etc. Each method may need different collaborators (the user API, a notification API, a search index, a logger). Threading these as separate parameters to every method creates 5–10-argument signatures that are painful to maintain and painful to add to. Storing them as fields on the request struct turns request types into stateful objects, which defeats the point of having them as immutable input contracts.

The cleanest answer is: **pass one thing** — the `*config.App` — and have the request method pull whatever it needs from `app.API().X` or `app.Logger()` at the call site. One parameter. Stable across time. Discoverable at the call site.

**Driver 2 — Cross-service calls should go through the target service's API, not its repository.**

If post's service could reach directly into user's repository (`app.Repos().User.Save(...)`), the user service would lose control over its own invariants. Any logging, validation, auditing, or business rule user wants to enforce on writes would be silently bypassed by every caller that knows the repo's shape. At scale this is a correctness disaster.

The guarantee we want is: **cross-service work always passes through the target service's interface**, so the target service gets to enforce its rules. The cleanest way to make this a structural guarantee (not a convention) is to **never expose repositories from the `App`**. Services receive their own repos via constructor injection; no other service can reach them because the `App` has no `Repos()` method to call.

**Driver 3 — The import graph must be a DAG even when services call each other in both directions.**

At runtime, a post operation may call user, and a user operation may call post. Go doesn't mind runtime circular calls (as long as they terminate), but it does mind circular **package imports**. If post's service package imports user's service package for the concrete type, and user's service package imports post's service package for the concrete type, that's a compile error.

The fix is the same one Go codebases have used forever: **interfaces in a stable leaf package**. Each service's root package (e.g. `post/`) declares the service's interface (`post.PostAPI`) and its request/view types. The service sub-package (`post/service/`) implements the interface. Cross-service calls import the **root package** for the interface type, never the service sub-package for the concrete type. The root is a stable leaf that imports nothing else in the project, so it breaks any potential cycle.

**Driver 4 — Testing should not get dramatically harder as the app grows.**

The naïve version of this pattern has a problem: to test any service that makes cross-service calls, you need to populate the `config.API` struct with working stubs for every service the code path touches. If you have 20 services, naïve tests become 40-line fixtures.

The solution is a shared `testapp.New()` helper plus per-API function-field stubs. Both are written once and reused. Per-test cost stays constant (3–5 lines) regardless of how many services exist. Details in [§13](#13-testing).

These four drivers, taken together, produce the pattern described below. Every specific rule traces back to one of them.

## 3. The `App` container

Defined in `backend/config/app.go`:

```go
type App struct {
    env            *provider.EnvProvider
    db             *sql.DB
    redis          *redis.Client
    rootDir        string
    logger         *slog.Logger
    queue          *provider.AsynqProvider
    sentry         *sentryhttp.Handler
    identityClient identity.Client

    // Cross-service service-API container. Populated by wiring code
    // (webserver.wireServices) after the App is constructed but before any
    // request is served.
    api *API
}
```

The `App` holds **infrastructure** (DB handle, logger, identity client, Redis, Sentry, queue) and **exactly one cross-service dispatch container** (`*API`). **It does not hold repositories.** See [§10](#10-repositories) for where they live instead, and [§2 Driver 2](#2-why-this-shape) for why.

Accessors follow a lazy-initialization pattern for the heavy infrastructure fields:

```go
func (app *App) DB() *sql.DB {
    if app.db == nil {
        // ... initialize from config
    }
    return app.db
}
```

The `API()` accessor is a simple field getter — no initialization — because the API container is populated once at startup by `webserver.wireServices`:

```go
func (app *App) API() *API {
    return app.api
}

func (app *App) RegisterAPI(api *API) {
    app.api = api
}
```

**Rule**: don't add fields to `App` that aren't genuinely cross-cutting infrastructure. Per-service state lives in that service's struct, not on `App`.

## 4. The `API` container

Defined in `backend/config/api.go`:

```go
type API struct {
    Post post.PostAPI
    User user.UserAPI
    // Add one field per service.
}
```

A plain struct with one field per service. Each field is typed as **the service's public API interface**, which lives in the service's root package (not the service sub-package — that's the cycle-breaking point, see [§12](#12-import-graph-and-why-its-acyclic)).

Access from any handler or service that holds `*config.App`:

```go
h.app.API().Post.CreatePost(ctx, postCreateRequest)
s.app.API().User.IncrementPostCount(ctx, incrementRequest)
```

Under the hood, each field holds a `*Service` pointer that implements the interface. Interface dispatch resolves to the concrete method at runtime. Callers only see the interface.

**When to add a new field**: whenever you add a new service with a public API that other services or handlers need to call. Adding a field is one line here and one line in `wireServices`. Nothing else in the app needs to know.

**When NOT to add fields**: internal helpers, private collaborators that only one service uses, or anything that isn't meant to be called from other services. Those live as struct fields on the consuming service, not on `config.API`.

## 5. Service layout

Every service lives under `backend/internal/app/services/<name>/` and follows the same directory structure. Domain types do NOT live per-service — they live in a **shared `backend/internal/app/domain/` package** that every service imports. Example from `backend/internal/app/services/post/`:

```
backend/internal/
├── app/                                 # Application domain + business logic
│   ├── apperrors/                       # Shared error vocabulary (see §16)
│   │   ├── errors.go                    #   Sentinels + Coded type + constructors
│   │   └── codes.go                     #   Translation key constants
│   ├── core/
│   │   └── actor/                       # Actor type (identity on ctx)
│   ├── domain/                          # ← shared domain package
│   │   ├── post.go                      #   type Post struct { ... }
│   │   ├── user.go                      #   type User struct { ... }
│   │   └── ...                          #   one file per entity type
│   ├── infrastructure/                  # Non-persistence external clients
│   │   └── identity/
│   └── services/
│       ├── post/
│       │   ├── api.go                   # Public contract: PostAPI interface, request types
│       │   ├── repository.go            # PostRepository interface (same package as api.go)
│       │   ├── permissions.go           # (Optional) permission constants
│       │   └── service/
│       │       ├── service.go           # *Service struct + New + compile-time check
│       │       ├── create_post.go       # (s *Service) CreatePost — one file per method
│       │       ├── get_post.go          # (s *Service) GetPost
│       │       ├── list_all_posts.go    # (s *Service) ListAllPosts
│       │       ├── list_posts.go        # (s *Service) ListPosts
│       │       ├── update_post.go       # (s *Service) UpdatePost
│       │       └── delete_post.go       # (s *Service) DeletePost
│       └── user/
│           ├── api.go
│           ├── repository.go
│           └── service/
│               ├── service.go           # *Service struct + New + compile-time check
│               ├── get_user.go          # (s *Service) GetUser
│               └── increment_post_count.go  # (s *Service) IncrementPostCount
├── webserver/                           # HTTP entry point (transport layer)
│   ├── handler/                         # Strict-server handlers
│   ├── middleware/                      # HTTP middleware (actor, CORS, logger, encoder)
│   └── webserver.go                     # Router + wireServices
└── worker/                              # Async worker entry point
    └── middleware/                      # Worker middleware
```

**The whole `internal/app/services/post/` directory is "the post service."** No more "module" terminology — the directory IS the service, and it contains the service's public API surface plus its implementation.

**Services return domain types directly.** There is no `PostView` / `UserView` / `XView` parallel type. `PostAPI.GetPost` returns `*domain.Post`. `PostAPI.ListAllPosts` returns `[]domain.Post`. If a field on the domain type should not be exposed to external callers, make the field **unexported** (lowercase) on the domain struct and add a public getter method — Go's package-level visibility enforces the boundary for free, without requiring a duplicate type. The View pattern is only worth introducing when you need a shape that spans multiple services (composed data), API versioning where the public contract diverges from the domain, or a read model with fundamentally different fields than the write model. In this codebase, none of those apply yet, so services return domain types.

Notes on each file:

- **`api.go`** — defines the exported `PostAPI` interface (method set only; no implementation) and the request types the interface uses. Imports only `context`, `internal/app/domain` (for the entity types the interface mentions), `internal/app/apperrors` (for the shared error vocabulary used in `Validate()` methods), and `internal/app/core/actor` if needed. **This file is the service's public surface.** Anyone importing `github.com/.../services/post` sees exactly these types.
- **`repository.go`** — `PostRepository` interface in the same package as `api.go`. Same package means the service's own repository contract is colocated with the rest of its public API.
- **`permissions.go`** — (optional) permission constants if the service uses them.
- **No per-service `errors.go`**. Error vocabulary is shared across every service via `internal/app/apperrors` (see §16). A service does not declare its own sentinel errors.
- **`service/`** — the concrete `Service` sub-package. Contains **one file per service method**, plus `service.go` which holds only the struct definition, the constructor, and the compile-time check. `service.go` never contains method bodies. Each method file is named after its method in `snake_case` (e.g. `create_post.go` holds `(s *Service) CreatePost`). Each file imports only what its method needs. Imports from `config` (for `*config.App`), its own service root (for types and the interface), `internal/app/domain` (for entity types), `internal/app/apperrors` (for error constructors), and any other services' **roots** (for cross-service calls). Never imports another service's `service/` sub-package.

**Why the `service/` sub-package exists**: it's the cycle-breaking seam. `config.API` imports `services/post` for the `PostAPI` interface type. `services/post/service` imports `config` for `*config.App`. If the implementation lived in `services/post` directly (same package as the interface), we'd have a cycle. Splitting into root-package-for-interface + sub-package-for-implementation is what keeps the import graph a DAG. See §12.

**Why domain is shared instead of per-service**: to avoid per-service domain packages importing each other. Cross-service pure functions (like the point-recalculation example from the design notes) can live in one `domain/` package where both entity types and their pure transformation functions have access to each other without import cycles. Entity ownership is conveyed through file naming (`domain/post.go` vs `domain/user.go`), not directory structure.

## 6. The `Service` struct

Every service has the same shape. From `backend/internal/app/services/post/service/service.go`:

```go
package service

import (
    "context"

    "github.com/noueii/no-frame-works/config"
    "github.com/noueii/no-frame-works/internal/app/services/post"
    "github.com/noueii/no-frame-works/internal/app/domain"
    "github.com/noueii/no-frame-works/internal/app/services/user"
)

type Service struct {
    app  *config.App          // for cross-service API calls via s.app.API().Other.X
    repo post.PostRepository  // for this service's own data access
}

func New(app *config.App, repo post.PostRepository) *Service {
    return &Service{app: app, repo: repo}
}

// Compile-time check that *Service satisfies post.PostAPI.
var _ post.PostAPI = (*Service)(nil)
```

**Two fields, always**: `app` and `repo`. The constructor takes both. This is the shape for every service in the codebase.

**Why both?**

- `repo` is the service's **own** data access. It's declared in the struct so the dependency is visible and so it can be swapped with a fake in tests directly (no need to thread it through the App).
- `app` is the **cross-service** dispatch and infrastructure container. The service reads `s.app.API().Other.X` for inter-service calls, `s.app.Logger()` for logging, `s.app.DB()` for raw DB access in rare cases, etc.

**The rule this shape enforces**: `s.app` has no `Repos()` accessor, so there is **no compile-time-valid way** for this service to reach another service's repository. The compiler forces cross-service data access through `s.app.API().Other.X`, which dispatches to the other service's implementation, which enforces the other service's invariants.

### Method bodies

Every service method is a linear script. Example — `CreatePost`:

```go
func (s *Service) CreatePost(ctx context.Context, req post.CreatePostRequest) (*domain.Post, error) {
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    newPost := domain.Post{
        Title:    req.Title,
        Content:  req.Content,
        AuthorID: req.AuthorID,
    }

    created, err := s.repo.Create(ctx, newPost)
    if err != nil {
        return nil, fmt.Errorf("failed to create post: %w", err)
    }

    // Cross-service call: keep the author's denormalized post count in sync.
    if err := s.app.API().User.IncrementPostCount(ctx, user.IncrementPostCountRequest{
        UserID: req.AuthorID,
    }); err != nil {
        return nil, fmt.Errorf("failed to increment author post count: %w", err)
    }

    return created, nil
}
```

Notice the shape:

1. `req.Validate()` — pure, no dependencies.
2. Build a `domain.X` from the request — pure.
3. `s.repo.Create(...)` — own data access via the struct field.
4. `s.app.API().Other.X(...)` — cross-service call via the App.
5. Build a `View` from the created domain object and return it.

Every service method is some variation of this pattern.

## 7. Cross-service calls

When post needs something from user (or vice versa), the call **must** go through `s.app.API().Other.X`. There is no other legal path.

```go
// INSIDE post.Service:

//  ✅ Allowed — go through user's public API.
err := s.app.API().User.IncrementPostCount(ctx, user.IncrementPostCountRequest{
    UserID: req.AuthorID,
})

//  ❌ Compile error — s.app has no Repos() method.
user, _ := s.app.Repos().User.FindByID(ctx, id)

//  ❌ Also a compile error — post.Service doesn't hold a user.UserRepository field.
user, _ := s.userRepo.FindByID(ctx, id)
```

The target service's implementation (`user.Service.IncrementPostCount`) runs the full validation, permission, and business rule pipeline before touching user state. Whatever user invariants exist will be enforced — because the call is going through the user service, not around it.

### The call chain, end to end

Here's what actually happens when `post.Service.CreatePost` calls `user.IncrementPostCount`:

```
post.Service.CreatePost(ctx, req)
  └─► s.app.API().User.IncrementPostCount(ctx, increment)
       └─► [interface dispatch resolves to *userservice.Service]
       └─► user.Service.IncrementPostCount(ctx, increment)
            └─► req.Validate()                              // validation
            └─► s.repo.IncrementPostCount(ctx, userID)       // user's own repo
                 └─► [user's persistence logic runs]
       [returns back up the chain]
  [post.Service.CreatePost continues]
```

Every hop is a normal Go function call. The only "magic" is the single interface dispatch at `s.app.API().User.X`, which resolves to the concrete `*userservice.Service`'s method table.

### Runtime cycles

Nothing prevents `A → B → A` call chains at runtime. Go doesn't care about runtime cycles as long as they terminate. If `user.MergeUsers` calls `post.ReassignPostsByAuthor` which calls `user.IncrementPostCount`, the chain is three hops and resolves cleanly.

**What to watch for**: unterminated recursion. If `user.IncrementPostCount` did something that eventually called `post.ReassignPostsByAuthor` again, which called `user.IncrementPostCount` again... that's an infinite loop, not a pattern problem. Prevent it by designing call graphs carefully, not by relying on the type system to catch it.

## 8. Request struct methods

Request types are immutable input contracts — never carry state across method calls. **Each request type lives in its own file at the service package root**, named after the operation: `services/post/create_post.go`, `services/post/get_post.go`, `services/user/increment_post_count.go`, etc. The file holds the struct **and its methods together**.

`api.go` holds **only the interface** — adding a new operation is one new sibling file plus one line in the interface, never an edit to a growing api.go.

```
services/post/
  api.go              // PostAPI interface only
  permissions.go      // permission constants
  repository.go       // PostRepository interface
  create_post.go      // CreatePostRequest + Validate + Run + Permission
  get_post.go         // GetPostRequest  + Validate + Run + Permission
  list_all_posts.go   // ListAllPostsRequest (empty) + Run
  list_posts.go       // ListPostsRequest + Validate + Run + Permission
  update_post.go      // UpdatePostRequest + Validate + Run
  delete_post.go      // DeletePostRequest + Validate + Run
  service/            // concrete *Service, one wrapper file per operation
    service.go
    create_post.go
    get_post.go
    ...
```

Why split this way: each operation has exactly two files — one in the package root (request type + per-operation logic) and one mirrored under `service/` (the thin wrapper that satisfies the interface and adds cross-service calls). When you read `create_post.go` you see the whole operation; you never have to scroll through a 300-line `api.go` to find one struct.

Every request type carries **three** kinds of methods, in order of how the service layer reaches for them:

1. `Validate()` — pure, no dependencies, called first.
2. `Run(ctx, repo)` — owns the single-service operation lifecycle.
3. `Permission()` (optional) — returns the static permission identifier for the operation.

### Pure methods: `Validate`

```go
func (r CreatePostRequest) Validate() error {
    if r.Title == "" {
        return apperrors.Validation(apperrors.CodePostTitleRequired, "title is required", nil)
    }
    if r.Content == "" {
        return apperrors.Validation(apperrors.CodePostContentRequired, "content is required", nil)
    }
    if r.AuthorID == "" {
        return apperrors.Validation(apperrors.CodePostAuthorIDRequired, "author_id is required", nil)
    }
    return nil
}
```

These test by calling them with literal values. No App, no fakes, no ceremony.

### `Run`: the request owns its lifecycle

Every request type has a `Run(ctx, repo)` method that owns the single-service operation: validate, build the domain object, touch the repo, return the result. The service method becomes a thin wrapper around `req.Run`.

```go
// services/post/api.go
func (r GetPostRequest) Run(ctx context.Context, repo PostRepository) (*domain.Post, error) {
    if err := r.Validate(); err != nil {
        return nil, errors.Errorf("post.GetPostRequest.Run: validate: %w", err)
    }
    found, err := repo.FindByID(ctx, r.ID)
    if err != nil {
        return nil, errors.Errorf("post.GetPostRequest.Run: repo find id=%s: %w", r.ID, err)
    }
    if found == nil {
        return nil, apperrors.NotFound(
            apperrors.CodePostNotFound, "post not found",
            map[string]any{"post_id": r.ID},
        )
    }
    return found, nil
}

// services/post/service/get_post.go
func (s *Service) GetPost(ctx context.Context, req post.GetPostRequest) (*domain.Post, error) {
    return req.Run(ctx, s.repo)
}
```

Why move the work onto the request? The service method only exists to satisfy the `PostAPI` interface and to layer in cross-service work. The actual per-operation logic lives next to the request type that defines the operation's input contract — same file, same package, same place a reader looks to understand "what does CreatePost do." Tests can drive `Run` directly with a fake repo, no `*Service` needed.

#### Why `Run(ctx, repo)` and not `Run(ctx, *config.App)`

`services/post/api.go` cannot import `config`. The reason is the same one that makes the whole pattern compile in the first place: `config` already imports this package for the `PostAPI` interface, so `post → config → post` is a direct cycle that Go rejects.

The workable alternative — and the one this codebase uses — is to give `Run` only its own service's repo:

```go
func (r CreatePostRequest) Run(ctx context.Context, repo PostRepository) (*domain.Post, error)
```

Run owns single-service work. Cross-service work stays in the service method, which has `s.app` in scope:

```go
// services/post/service/create_post.go
func (s *Service) CreatePost(ctx context.Context, req post.CreatePostRequest) (*domain.Post, error) {
    created, err := req.Run(ctx, s.repo)
    if err != nil {
        return nil, errors.Errorf("service.post.CreatePost: %w", err)
    }
    if err := s.app.API().User.IncrementPostCount(ctx, user.IncrementPostCountRequest{
        UserID: req.AuthorID,
    }); err != nil {
        return nil, errors.Errorf("service.post.CreatePost: increment author post count id=%s: %w", req.AuthorID, err)
    }
    return created, nil
}
```

The split is clean: **Run owns the single-service operation; the service method owns cross-service orchestration and anything else that needs `*config.App`.** If a Run method ever wants `app`, that's a sign the work belongs in the service method instead.

#### List operations get an empty request struct

Even parameterless list operations get a Run method, so every operation has the same shape:

```go
type ListAllPostsRequest struct{}

func (r ListAllPostsRequest) Validate() error { return nil }

func (r ListAllPostsRequest) Run(ctx context.Context, repo PostRepository) ([]domain.Post, error) {
    posts, err := repo.ListAll(ctx)
    if err != nil {
        return nil, errors.Errorf("post.ListAllPostsRequest.Run: repo list: %w", err)
    }
    return posts, nil
}
```

Handlers and callers pass `post.ListAllPostsRequest{}`. The cost is one empty struct; the benefit is no special-case parameterless methods on the API interface.

#### Methods on a loaded domain object stay pure

If a check depends on a domain object the service has already fetched (ownership checks, state transitions on an existing entity), keep it as a pure method on the request that takes the loaded object as a parameter:

```go
func (r UpdatePostRequest) CheckOwnership(actor actor.Actor, existing *domain.Post) error {
    if !existing.CanModify(actor) {
        return apperrors.Forbidden(apperrors.CodeUnauthorized, "cannot modify post", nil)
    }
    return nil
}
```

The service (or `Run`) fetches, then calls the pure method. Simpler to test, simpler to reason about.

**Test implication**: `req.Run(ctx, repo)` is testable with just a fake repo — no App, no service, no wiring. See [§13](#13-testing).

## 9. Handlers

Handlers live in `backend/internal/webserver/handler/`. They are **thin** — every handler has roughly the same shape:

```go
func (h *Handler) PostCreatePost(
    ctx context.Context,
    request oapi.PostCreatePostRequestObject,
) (oapi.PostCreatePostResponseObject, error) {
    a := actor.ActorFrom(ctx)
    if a == nil {
        return oapi.PostCreatePost400JSONResponse{
            ErrorJSONResponse: oapi.ErrorJSONResponse{Error: "unauthorized"},
        }, nil
    }

    result, err := h.app.API().Post.CreatePost(ctx, post.CreatePostRequest{
        Title:    request.Body.Title,
        Content:  request.Body.Content,
        AuthorID: a.UserID().String(),
    })
    if err != nil {
        return oapi.PostCreatePost400JSONResponse{
            ErrorJSONResponse: oapi.ErrorJSONResponse{Error: err.Error()},
        }, nil
    }

    return oapi.PostCreatePost201JSONResponse(toOAPIPost(result)), nil
}
```

Four responsibilities:

1. **Extract the actor** from context (for operations that need to know who's calling).
2. **Unwrap the oapi request** into the service's request type.
3. **Call `h.app.API().Post.CreatePost(...)`**.
4. **Map errors and successes** into oapi response types. Use `errors.Is` on sentinel errors to pick HTTP status codes.

The handler struct itself is minimal:

```go
type Handler struct {
    oapi.StrictServerInterface

    app      *config.App
    identity identity.Client
}

func NewHandler(app *config.App) *Handler {
    return &Handler{
        app:      app,
        identity: app.IdentityClient(),
    }
}
```

Two fields: the App and the identity client (used only for auth endpoints). No per-service fields. No wired-up service references. The handler reads `h.app.API().Post` on every call — which is cheap because it's a pointer dereference + a field read.

**Why no `postAPI` cache field?** Because caching it as a field adds zero value and subtracts one property: if the App's API is re-wired at runtime (in tests, or for feature-flag experiments), a cached field would miss the update. `h.app.API().Post` always sees the current registered API.

## 10. Repositories

Two concerns: where the interface lives, and where the implementation lives.

### Interface

The repository interface lives **in the service root package** (same package as `api.go`), in a file called `repository.go`:

```go
// backend/internal/app/services/post/repository.go
package post

import (
    "context"

    "github.com/noueii/no-frame-works/internal/app/domain"
)

type PostRepository interface {
    FindByID(ctx context.Context, id string) (*domain.Post, error)
    ListAll(ctx context.Context) ([]domain.Post, error)
    ListByAuthor(ctx context.Context, authorID string) ([]domain.Post, error)
    Create(ctx context.Context, post domain.Post) (*domain.Post, error)
    Update(ctx context.Context, post domain.Post) (*domain.Post, error)
    Delete(ctx context.Context, id string) error
}
```

The interface takes and returns **domain types** (`domain.Post`), not storage-model types. Mapping between domain and storage is a private concern of the concrete implementation.

### Concrete implementation

The Postgres implementation lives in `backend/repository/<service>/` as a separate package:

```go
// backend/repository/post/postgres.go
package post

import (
    "database/sql"

    postmod "github.com/noueii/no-frame-works/internal/app/services/post"
)

type PostgresPostRepository struct {
    db *sql.DB
}

func New(db *sql.DB) *PostgresPostRepository {
    return &PostgresPostRepository{db: db}
}

var _ postmod.PostRepository = (*PostgresPostRepository)(nil)
```

Each operation (`Create`, `FindByID`, etc.) gets its own file: `create.go`, `find_by_id.go`, `list_by_author.go`, etc. This is the "one exported function per file" convention we use for repository layers.

Internally, each method uses go-jet's query builder, maps to/from the storage model via private `toModel` / `toDomain` helpers, and returns domain types.

### Not on the `App`

The concrete repository is **constructed by the wiring function** and passed directly into the service constructor that owns it. The App never holds a reference to it.

```go
// in webserver.wireServices:
pRepo := postrepo.New(app.DB())              // local variable, not stored on app
pSvc  := postservice.New(app, pRepo)          // passed to the service's constructor
```

See [§11](#11-wiring) for the full wiring flow.

## 11. Wiring

All wiring happens in **one function**: `webserver.wireServices` in `backend/internal/webserver/webserver.go`. This is deliberate — keeping every construction in one file means dependency order is visible and mistakes are obvious.

```go
func wireServices(app *config.App) {
    // Repositories — local variables only, never stored on the App.
    pRepo := postrepo.New(app.DB())
    uRepo := userrepo.New(app.DB())

    // Services — each takes the App (for cross-service API access via
    // app.API()) and its own repository as a directly injected field.
    // Services cannot reach each other's repositories.
    pSvc := postservice.New(app, pRepo)
    uSvc := userservice.New(app, uRepo)

    // Register the API container. After this line,
    // app.API().Post.CreatePost and app.API().User.IncrementPostCount are
    // callable from any handler or any other service that holds *config.App.
    app.RegisterAPI(&config.API{
        Post: pSvc,
        User: uSvc,
    })
}
```

Called from `NewWebserver` before the handler is constructed:

```go
func NewWebserver(app *config.App) *Webserver {
    wireServices(app)   // ← populates app.API() and all its fields

    h := handler.NewHandler(app)
    // ... rest of webserver setup
}
```

**Order of operations inside `wireServices`**:

1. **Build each repository** as a local variable. Repos only depend on `app.DB()`.
2. **Build each service** by calling `New(app, repo)` with the matching repo. Services hold `*config.App` and their own repo.
3. **Register the API container.** After this line, `app.API()` returns a non-nil struct and every service is reachable via its interface.

**Important**: the service constructors run while `app.API()` is still nil. As long as no constructor *dereferences* `app.API()` during construction, this is fine — and none should. The `New` function just stores the pointer. Cross-service calls via `s.app.API()` happen at request time, long after wiring finishes.

**If you ever need a service to do work during construction**, extract that work into an `Init()` method that's called after `RegisterAPI` — not into `New()`.

### Adding a new service

Three edits:

1. Add a field to `config.API` in `backend/config/api.go`.
2. Add ~2 lines to `wireServices`: construct the repo, construct the service, add the field to the `config.API` literal.
3. That's it. Every handler that already holds `*config.App` can now call `h.app.API().NewService.X`.

## 12. Import graph and why it's acyclic

The rule we rely on: **service root packages are stable leaves**. They import only `context`, their own `domain/` package, and (rarely) `core/actor`. They do **not** import `config`, the service sub-package, or any other service.

Service sub-packages can import:

- `config` (for `*config.App`)
- Their own service root (for request types, interfaces, sentinel errors)
- Their own `domain/` (for domain types)
- **Other services' roots** (for cross-service request types and interfaces) — but **never other services' sub-packages**.

This means the graph looks like:

```
┌───────────────┐
│ post (root)   │ ◄──── no internal dependencies besides context + domain
└───────┬───────┘
        │
┌───────▼───────┐         ┌───────────────┐
│ post/service  │────────►│ config        │
│               │         │               │
│               │────────►│ user (root)   │ ◄──── also a leaf
└───────────────┘         └───────┬───────┘
                                  │
                          ┌───────▼───────┐
                          │ user/service  │
                          │               │
                          └───────┬───────┘
                                  │
                          (imports post root for
                          cross-service calls to post)
```

**Key observation**: `post/service` imports `user` (the root), not `user/service` (the implementation). `user/service` imports `post` (the root), not `post/service`. The two services never directly import each other. The interfaces in the root packages are the seam that breaks potential cycles.

**What happens if you break this rule**: if `post/service` tried to `import "github.com/.../user/service"` to use the concrete `*userservice.Service` type, the graph would cycle (`post/service → user/service → config → post → ???`), and Go would reject the build at compile time.

**The compile error is your safety net.** If you ever get `import cycle not allowed` and the cycle involves two service sub-packages, the fix is: replace the direct service import with the root package import and use the API interface.

## 13. Testing

Two infrastructure pieces, both written once:

### Infrastructure piece 1: function-field stubs per API

For each service API, write a stub struct with one function field per interface method:

```go
// backend/internal/app/services/post/stub/stub.go
package stub

import (
    "context"

    "github.com/noueii/no-frame-works/internal/app/domain"
    "github.com/noueii/no-frame-works/internal/app/services/post"
)

type PostAPI struct {
    CreatePostFn   func(context.Context, post.CreatePostRequest) (*domain.Post, error)
    GetPostFn      func(context.Context, post.GetPostRequest) (*domain.Post, error)
    UpdatePostFn   func(context.Context, post.UpdatePostRequest) (*domain.Post, error)
    DeletePostFn   func(context.Context, post.DeletePostRequest) error
    ListAllPostsFn func(context.Context) ([]domain.Post, error)
    ListPostsFn    func(context.Context, post.ListPostsRequest) ([]domain.Post, error)
}

var _ post.PostAPI = (*PostAPI)(nil)

func (s *PostAPI) CreatePost(ctx context.Context, req post.CreatePostRequest) (*domain.Post, error) {
    if s.CreatePostFn == nil {
        return nil, nil
    }
    return s.CreatePostFn(ctx, req)
}
// ... one method per interface method, same shape
```

~70 lines per 6-method interface. **Written once**, not per test. Can be auto-generated with `moq` or `mockery` if you want to skip the hand-rolling.

The same pattern applies to `post.PostRepository`, `user.UserAPI`, `user.UserRepository`, and every other interface you want to mock.

### Infrastructure piece 2: the `testapp.New()` helper

```go
// backend/config/testapp/testapp.go
package testapp

import (
    "io"
    "log/slog"

    "github.com/noueii/no-frame-works/config"
    poststub "github.com/noueii/no-frame-works/internal/app/services/post/stub"
    userstub "github.com/noueii/no-frame-works/internal/app/services/user/stub"
)

// New returns a *config.App populated with safe default stubs for every
// service API and a silent logger. Individual tests override the function
// fields they care about.
func New() *config.App {
    app := &config.App{}
    // ... set logger to a silent slog for tests ...
    app.RegisterAPI(&config.API{
        Post: &poststub.PostAPI{},
        User: &userstub.UserAPI{},
    })
    return app
}
```

**Written once**. Updated one line per new service. That's all.

### Per-test setup

Three patterns, depending on what the test needs:

**Pattern A — test pure functions (no setup at all)**:

```go
func TestCreatePostRequest_Validate(t *testing.T) {
    err := post.CreatePostRequest{Title: "", Content: "c", AuthorID: "u1"}.Validate()
    if !errors.Is(err, post.ErrTitleRequired) {
        t.Fatalf("expected ErrTitleRequired, got %v", err)
    }
}
```

No App, no fakes, nothing. Just call the pure method with literal inputs. Use this for `Validate()`, domain predicates like `Post.CanModify`, and any other pure logic.

**Pattern B — test a service method with happy-path defaults (fake repos + real services)**:

```go
func TestPostService_CreatePost_HappyPath(t *testing.T) {
    postRepo := &poststub.Repo{}
    userRepo := &userstub.Repo{}

    app := &config.App{}
    userSvc := userservice.New(app, userRepo)
    postSvc := postservice.New(app, postRepo)
    app.RegisterAPI(&config.API{
        Post: postSvc,
        User: userSvc,
    })

    _, err := postSvc.CreatePost(ctx, post.CreatePostRequest{
        Title: "t", Content: "c", AuthorID: "u1",
    })
    // assertions on postRepo.created, userRepo.incremented, etc.
}
```

Real services, fake repos. Exercises the full cross-service chain end-to-end. This is the default test shape when you want to verify that everything wires correctly and the happy path works.

**Pattern C — test an error path (direct API stub)**:

```go
func TestPostService_CreatePost_WhenUserIncrementFails(t *testing.T) {
    app := testapp.New()

    // Override the specific method we're testing the error path of:
    app.API().User.(*userstub.UserAPI).IncrementPostCountFn = func(
        _ context.Context, _ user.IncrementPostCountRequest,
    ) error {
        return user.ErrUserNotFound
    }

    svc := postservice.New(app, &poststub.Repo{})

    _, err := svc.CreatePost(ctx, post.CreatePostRequest{
        Title: "t", Content: "c", AuthorID: "u1",
    })

    if !errors.Is(err, user.ErrUserNotFound) {
        t.Fatalf("expected ErrUserNotFound, got %v", err)
    }
}
```

Direct stub, one-liner override to force the error. 5 lines of setup. Scales to any edge case.

### Why per-test cost doesn't scale with service count

The key property of this testing pattern is that **adding a new service to the project does not change the per-test cost**. Each test already populates only the subset of `config.API` it actually touches (or uses `testapp.New()` for defaults). Adding a 15th service means one more field on `config.API`, one more line in `testapp.New()`, and one more stub file — all one-time costs. The 500 tests you already have don't change.

The only way per-test cost grows is if a specific test needs to stub more methods because the code path under test has grown more cross-service calls. That's proportional to the code path's complexity, not to the project's service count.

## 14. The hard rules

The non-negotiable rules, distilled. Breaking any of these is a review blocker:

1. **Services always have the shape `type Service struct { app *config.App; repo ModuleRepository }`.** Two fields, constructor takes both, `var _ ModuleAPI = (*Service)(nil)` compile-time check. No extra dependency fields on the struct.
2. **Never import another service's `service` sub-package.** Only import the root package (which contains the interface and request types). If you hit an import cycle, you've broken this rule.
3. **Cross-service state changes go through `s.app.API().Other.X(...)`.** Not through repositories. Not through direct struct references. The `App` has no `Repos()` method, so this is compiler-enforced.
4. **Request struct methods are pure, or take `(ctx, *config.App[, *domain.X])`.** Never store state on the request struct between method calls. Request types are data, not objects.
5. **Handlers hold `*config.App` and nothing else** (except `identity.Client` for auth endpoints). Every handler method reads `h.app.API().Post.X` per call.
6. **Wiring lives only in `webserver.wireServices`.** No per-service wiring, no hidden `init()` functions, no handler-local wiring. One place, one flow.
7. **Repositories return domain types, not storage-model types.** `FindByID` returns `*domain.Post`, not `*model.Post`. Mapping is private to the repository package.
8. **The service root (`post/`, `user/`) is a stable leaf.** It imports only `context`, its own `domain/`, and (rarely) `core/actor`. It never imports `config`, service sub-packages, or other services.

## 15. Adding a new service

Follow this checklist when adding a new service (e.g. `comment`):

1. **Add the entity type** to `backend/internal/app/domain/comment.go`: `type Comment struct { ... }` in `package domain`. If the type needs pure methods (like `CanModify`), add them here.
2. **Create the service directory**: `backend/internal/app/services/comment/`.
3. **Write `api.go`** with:
   - `CommentAPI` interface (method signatures, returning `*domain.Comment` for single reads and `[]domain.Comment` for lists).
   - Request types (e.g. `CreateCommentRequest`) with a `Validate()` method each.
4. **Write `errors.go`** with service-level sentinel errors.
5. **Write `repository.go`** with the `CommentRepository` interface (takes and returns `domain.Comment`).
6. **Write `service/service.go`** with just the `Service` struct (two fields: `app`, `repo`), the `New(app, repo)` constructor, and the compile-time check `var _ comment.CommentAPI = (*Service)(nil)`. **Do not put method bodies in `service.go`**.
7. **Write one file per service method** in `service/` — e.g. `create_comment.go` for `(s *Service) CreateComment(...)`, `get_comment.go` for `(s *Service) GetComment(...)`, and so on. The file name is the method name in `snake_case`. Each file imports only what that method's body needs. This keeps files small, makes navigation trivial (grep the method name as a file name), and keeps the import lists tight per-operation.
8. **Write `backend/repository/comment/postgres.go`** with the concrete `PostgresCommentRepository` implementing `comment.CommentRepository`. One file per operation: `create.go`, `find_by_id.go`, etc.
9. **Add a field to `config.API`** in `backend/config/api.go`: `Comment comment.CommentAPI`.
10. **Add wiring** in `webserver.wireServices`:
    ```go
    cRepo := commentrepo.New(app.DB())
    cSvc  := commentservice.New(app, cRepo)
    ```
    and add `Comment: cSvc,` to the `config.API` literal.
11. **Add handlers** in `backend/internal/webserver/handler/` for each oapi endpoint. Each handler calls `h.app.API().Comment.X(...)`.
12. **Write a stub** in `backend/internal/app/services/comment/stub/stub.go` (~70 lines, one function field per method).
13. **Update `testapp.New()`** in `backend/config/testapp/testapp.go` to register the comment stub by default.

Most of these steps are mechanical and take 5–10 minutes per service. The thinking happens in steps 1 (what is the entity and what are its invariants?), 3 (what does the service expose?), and 7 (what does each operation actually do?). Everything else is boilerplate you write once and don't revisit.

## 16. Errors

The application uses a **small, shared error vocabulary** plus a **typed `*Coded` error** for user-facing errors that need frontend translation. All error handling goes through one package: `backend/internal/app/apperrors`.

### 16.1 The six sentinels

Every error the app produces eventually resolves (via `errors.Is`) to one of exactly six categorical sentinels. They map one-to-one to HTTP status codes in handlers:

| Sentinel | HTTP status | Meaning |
|----------|-------------|---------|
| `apperrors.ErrNotFound` | 404 | A requested entity does not exist |
| `apperrors.ErrValidation` | 400 | Input shape is wrong (field required, format invalid) |
| `apperrors.ErrUnauthorized` | 401 | No valid actor on context (not authenticated) |
| `apperrors.ErrForbidden` | 403 | Authenticated but lacks permission for this operation |
| `apperrors.ErrConflict` | 409 | State conflict (uniqueness violation, optimistic lock, etc.) |
| `apperrors.ErrInternal` | 500 | Anything else — infrastructure failures, bugs, panics |

That's the whole vocabulary. No per-service `ErrPostNotFound`, `ErrUserNotFound`, `ErrTitleRequired`, etc. The six sentinels cover every HTTP response the app will ever return.

### 16.2 The `*Coded` typed error

For errors the frontend needs to translate, wrap the sentinel in a `*apperrors.Coded`:

```go
type Coded struct {
    Code    string         // stable translation key, e.g. "post.title_required"
    Message string         // English fallback for logs and fallback UI
    Params  map[string]any // interpolation values, e.g. {"username": "alice"}
    Kind    error          // one of the six sentinels
}
```

The handler extracts the `Code` and `Params` from the error chain via `errors.As`, and passes them to the frontend in the response body. The frontend uses `Code` as a key in its translation file (i18next, FormatJS, etc.), interpolating `Params` into the translated string.

`Coded` also satisfies `errors.Is(ce, sentinel)` — it forwards to its `Kind` field. So handler HTTP-status matching is unchanged: `errors.Is(err, apperrors.ErrValidation)` still picks the 400 branch, regardless of whether the error is a plain sentinel or a `*Coded` wrapping it.

### 16.3 Constructors

Instead of building `*Coded` structs by hand, use the helpers:

```go
apperrors.Validation(code, message, params)    // wired to ErrValidation → 400
apperrors.NotFound(code, message, params)      // wired to ErrNotFound → 404
apperrors.Conflict(code, message, params)      // wired to ErrConflict → 409
apperrors.Forbidden(code, message, params)     // wired to ErrForbidden → 403
apperrors.Unauthorized(code, message, params)  // wired to ErrUnauthorized → 401
```

Each constructor pre-fills `Kind` with the correct sentinel so you can't accidentally mismatch category and helper.

### 16.4 Code constants

Error codes are declared as string constants in `backend/internal/app/apperrors/codes.go`:

```go
const (
    CodePostTitleRequired    = "post.title_required"
    CodePostContentRequired  = "post.content_required"
    CodePostNotFound         = "post.not_found"
    CodeUserNotFound         = "user.not_found"
    CodeUserIDRequired       = "user.id_required"
    // ...
)
```

Constants (not inline strings) so that typos are compile errors and a single file shows every code the app can produce. The values match the frontend translation keys one-to-one — changing a constant value requires updating every frontend translation file that defines it.

Naming convention: `<service>.<specific_reason>` — lowercase, dot-separated, underscores within segments. Group codes for the same service in the same block.

### 16.5 Writing a validation error in a service

On a request struct's `Validate()` method — the canonical place for shape validation:

```go
func (r CreatePostRequest) Validate() error {
    if r.Title == "" {
        return apperrors.Validation(apperrors.CodePostTitleRequired, "title is required", nil)
    }
    if r.Content == "" {
        return apperrors.Validation(apperrors.CodePostContentRequired, "content is required", nil)
    }
    return nil
}
```

The service method wraps the result with layer context:

```go
func (s *Service) CreatePost(ctx context.Context, req post.CreatePostRequest) (*domain.Post, error) {
    if err := req.Validate(); err != nil {
        return nil, errors.Errorf("service.post.CreatePost: validate: %w", err)
    }
    // ...
}
```

Note `errors.Errorf` is imported from `github.com/go-errors/errors` — not `fmt.Errorf`. The go-errors variant is `%w`-compatible with stdlib and additionally captures stack traces on wrap.

### 16.6 Writing a not-found error in a service

```go
func (s *Service) GetPost(ctx context.Context, req post.GetPostRequest) (*domain.Post, error) {
    if err := req.Validate(); err != nil {
        return nil, errors.Errorf("service.post.GetPost: validate: %w", err)
    }

    found, err := s.repo.FindByID(ctx, req.ID)
    if err != nil {
        return nil, errors.Errorf("service.post.GetPost: repo find id=%s: %w", req.ID, err)
    }
    if found == nil {
        return nil, apperrors.NotFound(
            apperrors.CodePostNotFound,
            "post not found",
            map[string]any{"post_id": req.ID},
        )
    }
    // ...
}
```

Three error return shapes in one method:

1. **Validation passthrough** — `req.Validate()` already returns a `*Coded`; the service wraps with its own layer context.
2. **Infrastructure error** — the repo returned an error; the service wraps with layer context and the error propagates as an uncategorized internal error (handler maps to 500).
3. **Explicit not-found** — the repo returned `nil, nil`; the service constructs a `*Coded` backed by `ErrNotFound` with a translation code and the post ID in params.

### 16.7 The layer prefix convention

Every `errors.Errorf` wrap starts with `<layer>.<service>.<operation>` so the error chain is self-labeling:

- **Repository**: `errors.Errorf("repo.post.FindByID id=%s: %w", id, err)`
- **Service**: `errors.Errorf("service.post.CreatePost: load existing: %w", err)`
- **Domain** (for pure functions that return errors): `errors.Errorf("domain.post.Validate: title required: %w", apperrors.ErrValidation)`
- **Handlers don't wrap** — they match sentinels and respond.

Reading a full chain left-to-right tells you the layer boundaries and the call path:

```
service.post.UpdatePost: load existing id=abc123: repo.post.FindByID: query post_table: connection refused
```

Grepping logs for `service.post.` finds every service-layer error in the post service. Grepping for `repo.post.FindByID` finds every call site of that specific method that failed.

### 16.8 Handler pattern

Handlers do three things: match the sentinel to pick an HTTP status, extract the `*Coded` for the response body, and log the full chain on unmatched errors.

```go
func (h *Handler) GetPost(ctx context.Context, request oapi.GetPostRequestObject) (oapi.GetPostResponseObject, error) {
    result, err := h.app.API().Post.GetPost(ctx, post.GetPostRequest{
        ID: request.Id.String(),
    })
    if err != nil {
        if errors.Is(err, apperrors.ErrNotFound) {
            return oapi.GetPost404JSONResponse{
                ErrorJSONResponse: oapi.ErrorJSONResponse{
                    Error: apperrors.Message(err, "post not found"),
                },
            }, nil
        }
        h.app.Logger().ErrorContext(ctx, "get post failed",
            slog.String("post_id", request.Id.String()),
            slog.String("error_code", apperrors.CodeOf(err)),
            slog.Any("error", err),
        )
        return nil, err
    }
    return oapi.GetPost200JSONResponse(toOAPIPost(result)), nil
}
```

- **`errors.Is(err, apperrors.ErrNotFound)`** — walks the chain, matches even when wrapped multiple layers deep.
- **`apperrors.Message(err, fallback)`** — extracts `Coded.Message` from the chain if present, else returns the fallback string.
- **`apperrors.CodeOf(err)`** — extracts `Coded.Code` for logging; empty string if the error is uncoded (infrastructure failure).
- **`slog.Any("error", err)`** — logs the full wrap chain with stack trace (go-errors captures one at each wrap site).

### 16.9 Response shape today vs. the full translation-ready shape

Today's `oapi.ErrorJSONResponse` only has an `Error` string field. Handlers populate it with `apperrors.Message(err, fallback)`, so the user-facing message from a `*Coded` makes it to the frontend — but `Code` and `Params` do not.

To expose `Code` and `Params` to the frontend for full i18n translation, update `openapi/shared.yaml` (or the bundled spec) to include them in the error response schemas, then run `make gen-openapi` to regenerate. The schema already has `errorCode`/`errorMessage`/`data` fields defined in `shared.yaml` — the generated Go code just hasn't been refreshed to pick them up. Regenerating will produce new handler fields that can be populated from `apperrors.CodeOf(err)` and `apperrors.ParamsOf(err)`.

This is a **separate follow-up step**, not required for the current refactor. The backend produces `*Coded` errors today; the response-body surface can be extended when the frontend is ready to consume them.

### 16.10 What NOT to do

- **Don't create per-service sentinel errors** like `post.ErrPostNotFound`. Use `apperrors.NotFound(CodePostNotFound, ...)` instead — it satisfies `errors.Is(err, apperrors.ErrNotFound)` for status mapping AND carries a translation code.
- **Don't use `fmt.Errorf`** for wrapping. Use `errors.Errorf` from `github.com/go-errors/errors` — it captures stack traces and is `%w`-compatible.
- **Don't use stdlib `errors.New`** for sentinel declarations. Use `errors.Errorf` or `errors.New` from go-errors.
- **Don't inline error codes** as strings in service bodies. Use the constants from `apperrors/codes.go` so typos are compile errors.
- **Don't log errors at multiple layers.** Log once, at the handler, where the error exits the program boundary. Layers below wrap and return.
- **Don't return bare sentinels from services.** Wrap with `errors.Errorf("service.X.Y: context: %w", err)` so the error chain has layer context. `errors.Is` still matches through the wraps.
- **Don't swallow errors.** Never write `_ = thing.DoSomething()` silently. Either return the error up the stack or log it explicitly.

### 16.11 Rule summary

1. **Six sentinels, period.** `ErrNotFound`, `ErrValidation`, `ErrUnauthorized`, `ErrForbidden`, `ErrConflict`, `ErrInternal`. That's the vocabulary.
2. **`*Coded` for user-facing errors** — carries a stable translation key, English fallback, and interpolation params.
3. **Wrap with `errors.Errorf` at every layer** using the `layer.service.operation: context: %w` convention.
4. **Handler matches sentinels** with `errors.Is` to pick HTTP status. Extracts `Code` and `Params` via helpers for the response body.
5. **Log once, at the handler**, with structured slog attributes for request ID, actor, and the full error chain.
6. **Use `github.com/go-errors/errors`** everywhere — `errors.New`, `errors.Errorf`, `errors.Is`, `errors.As`.

## 17. Known gaps and future work

Things the current pattern does **not** address. These are conscious omissions — the pattern works without them, and they can be added later without changing the shape.

- **No transaction manager.** A service method that writes to two different repos (or that calls another service that also writes) has no way to make those writes atomic. Adding a `shared.TxManager` with ctx-based propagation is documented in [`08-transactions.md`](08-transactions.md); it can be added later without changing any service struct.
- **No permission middleware.** The `PermissionLayer` wrapper that used to sit in front of `post.Service` was dropped. Authorization is currently ad-hoc (inline in handlers via `actor.ActorFrom(ctx)`). If centralized auth becomes important, you can either (a) re-introduce a per-service middleware that wraps the service in the `config.API` registration, or (b) add a `CheckPermission(ctx, *config.App)` method to each request struct and call it at the top of every service method.
- **No nil-safety on `app.API()`.** If anyone manages to call `app.API()` before `wireServices` has run, they get a nil-pointer panic with no useful message. Add a check in the accessor if this becomes a real risk:
    ```go
    func (app *App) API() *API {
        if app.api == nil {
            panic("config.App.API() called before RegisterAPI; ensure webserver.wireServices ran")
        }
        return app.api
    }
    ```
- **The stub files and `testapp.New()` helper don't exist yet**. They are described in [§13](#13-testing) but haven't been added to the codebase. They should be added as the test suite grows, before you have more than a few tests that need fakes.
- **No observability middleware.** Tracing, metrics, structured logging with request IDs — none of these are wired uniformly. They'll eventually want a thin wrapper around services (similar to the dropped `PermissionLayer`) or inline calls to `s.app.Logger()`.
- **`identity.Client` is a one-off field on `Handler`**. If more infra clients accumulate there (e.g. an email sender, a feature-flag client), they should move into `*config.App` and be read via `app.X()` from wherever they're used.

None of these gaps block shipping the current pattern. Address them when they become pain points, not preemptively.

---

## Appendix: the "why god-App" decision, restated

We considered and rejected these alternatives:

- **Constructor injection without god-App.** Services take every dependency as a constructor parameter. Clean for small services, but fails when request-scoped methods need access to 5+ collaborators. Signature bloat becomes the dominant cost.
- **Dependency struct per service.** `type Deps struct { Repo X; UserAPI Y; ... }` passed to `New(Deps{...})`. Nice ergonomics, but doesn't help with request methods that need diverse dependencies — you'd still have to thread the `Deps` struct into every request method, or duplicate it as state on the request.
- **Narrow context interfaces per service.** Each service declares its own `AppContext` interface listing only the methods it uses from the App. Gives compile-time scoping but adds significant boilerplate (one interface per service, kept in sync with the App's methods).
- **DI framework (wire, fx).** Auto-generates wiring from provider functions. Powerful at scale but introduces a code-gen step and obscures the construction flow. Worth considering if service count grows beyond ~20.

The god-App with compile-time repository isolation wins because it optimizes for two specific axes:

1. **Request-scoped methods with diverse dependencies** — cheap access through `s.app.API()` regardless of how many collaborators the method uses.
2. **Cross-service boundary enforcement** — compile-time guarantee that services can't reach into each other's storage, so target services always get to enforce their own invariants.

The cost — heavier test fixtures — is manageable with the `testapp.New()` helper, and it doesn't grow with project size. At the scale this template is meant to support (small team, 10–20 services), the tradeoff is clearly favorable.
