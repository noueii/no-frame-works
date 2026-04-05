# Backend Framework Guide

A modular monolith framework in Go. This document defines the conventions, patterns, and rules for building backend modules. It is project-agnostic — any application built on this framework follows these rules.

Examples throughout this document use a Japanese learning app (japcom) for illustration. Substitute your own module and type names.

## Project Structure

```
backend/
├── cmd/
│   ├── webserver/main.go          # Wires all modules, starts HTTP server
│   └── worker/main.go             # Background job runner
├── config/
│   └── app.go                     # Dependency injection container
├── db/
│   └── migrations/                # Goose SQL migrations
├── generated/                     # OpenAPI generated types (oapi-codegen)
├── internal/
│   ├── core/                      # Cross-cutting infrastructure
│   │   └── actor/                 # Actor system (see Actor System section)
│   ├── modules/                   # Business modules
│   │   ├── <module_a>/
│   │   ├── <module_b>/
│   │   └── <module_c>/
│   └── provider/                  # External system adapters
│       ├── <provider_a>/          # Interface + implementations
│       └── <provider_b>/          # Interface + implementations
└── repository/                    # Concrete repository implementations
    ├── <module_a>/
    ├── <module_b>/
    └── <module_c>/
```

## Module Structure

Every module follows this structure:

```
modules/<name>/
├── api.go                         # Public API interface + exported types
├── repository.go                  # Repository interface
├── permissions.go                 # Permission constants
├── handler/
│   └── http/
│       ├── handler.go             # Handler struct, constructor
│       ├── routes.go              # RegisterRoutes(router)
│       └── <verb>_<resource>.go   # One file per endpoint
├── service/
│   ├── service.go                 # Service struct, implements API interface
│   └── <concern>/
│       └── <function_name>.go     # One file per function
├── domain/
│   ├── models.go                  # Module-specific types
│   ├── enums.go                   # Module-specific enums
│   └── errors.go                  # Module-specific errors
└── middleware/
    └── permission.go              # Permission wrapper around API interface
```

### Rules

- Modules never import other modules directly. Cross-module communication goes through API interfaces.
- Each module owns its types. No shared domain package. Convert at boundaries.
- `handler/http/` is the HTTP entry point. The `http/` subfolder allows adding `grpc/`, `ws/`, etc. later.
- Service subfolders are separate Go packages. One function per file.
- Handler files are named `<verb>_<resource>.go` (e.g., `get_items.go`, `post_create.go`).

## Module API (api.go)

The API interface is the public contract of a module. Other modules depend on this interface, never on the service or repository directly.

```go
package orders

import "context"

// OrdersAPI is the public contract for the orders module.
type OrdersAPI interface {
    CreateOrder(ctx context.Context, req CreateOrderRequest) (OrderResult, error)
    GetOrder(ctx context.Context, req GetOrderRequest) (OrderView, error)
}

// Exported types used in the API contract.
// These are what other modules receive when calling this API.

type OrderView struct {
    ID     int
    Status string
    Total  float64
}
```

### Rules

- The API interface is defined in `api.go` at the module root.
- Types returned by the API are defined in `api.go` alongside the interface. These are the "view" types that external consumers see.
- Internal types (used only within the module) live in `domain/`.
- The service implements this interface. The permission middleware wraps it.

## Repository Pattern

### Interface (module-owned)

Each module defines its repository interface in `repository.go`:

```go
package orders

import "context"

type OrdersRepository interface {
    FindByID(ctx context.Context, id int) (*domain.Order, error)
    Create(ctx context.Context, order domain.Order) error
}
```

### Implementation (separate package)

Concrete implementations live in the top-level `repository/` directory, split into subfolders with one function per file:

```
repository/<module_name>/
├── postgres.go                    # Struct, constructor, implements interface
├── <concern_a>/
│   ├── find_by_id.go
│   └── create.go
└── <concern_b>/
    ├── list.go
    └── update.go
```

```go
package orders

import "database/sql"

type PostgresOrdersRepository struct {
    db *sql.DB
}

func New(db *sql.DB) *PostgresOrdersRepository {
    return &PostgresOrdersRepository{db: db}
}
```

### Rules

- The module owns the interface. The repository package provides the implementation.
- Repository subfolders are separate Go packages. One function per file.
- Repository implementations use Jet for type-safe SQL queries.
- The module never imports the concrete repository — it depends on the interface. Wiring happens in `cmd/webserver/main.go`.

## Validation

Input validation uses self-validating request structs. Every request struct passed to an API method must implement `Validatable`:

```go
type Validatable interface {
    Validate() error
}
```

### Request Struct Example

```go
type CreateOrderRequest struct {
    UserID    int
    ProductID int
    Quantity  int
}

func (r CreateOrderRequest) Validate() error {
    if r.ProductID == 0 {
        return ErrProductRequired
    }
    if r.Quantity <= 0 {
        return ErrInvalidQuantity
    }
    return nil
}
```

### Where Validation Happens

| Layer | What it validates |
|-------|-------------------|
| **Request struct** (`Validate()`) | Input shape: required fields, valid enums, format, ranges |
| **Service logic** | Domain rules: does the item exist, is the state valid, does the user have access |

The service calls `req.Validate()` as its first operation. Domain validation happens naturally during business logic execution.

### Rules

- All request structs implement `Validate() error`.
- `Validate()` only checks input shape — it never queries the database.
- Domain validation (existence checks, state checks) lives in the service.
- Handlers parse HTTP input into request structs but do not validate — that is the service's responsibility.

## Permission System

Permissions are enforced via a middleware wrapper around the module's API interface. The system has two components: permission declarations on request structs and a permission layer that checks them.

### Permission Declaration

Every request struct declares what permission it requires by implementing `Authorizable`:

```go
type Authorizable interface {
    Permission() Permission
}
```

```go
// permissions.go
const (
    PermOrderCreate Permission = "orders:order:create"
    PermOrderView   Permission = "orders:order:view"
    PermOrderCancel Permission = "orders:order:cancel"
)

// On the request struct
func (r CreateOrderRequest) Permission() Permission {
    return PermOrderCreate
}
```

### Permission Convention

Permissions follow the format: `<module>:<resource>:<action>`.

```
orders:order:create
orders:order:view
orders:order:cancel
inventory:product:view
inventory:stock:update
users:settings:update
users:profile:view
```

### Permission Layer (middleware/permission.go)

The permission layer wraps the API interface and checks authorization before forwarding to the service:

```go
type PermissionLayer struct {
    inner OrdersAPI
}

func (p *PermissionLayer) CreateOrder(ctx context.Context, req CreateOrderRequest) (OrderResult, error) {
    if err := p.authorize(ctx, req.Permission()); err != nil {
        return OrderResult{}, err
    }
    return p.inner.CreateOrder(ctx, req)
}
```

### Rules

- Every request struct implements both `Validatable` and `Authorizable`.
- Permission constants are defined in the module's `permissions.go`.
- The permission layer wraps the API interface — it is the outermost layer before the service.
- Adding a new API method without a corresponding permission on its request struct is a compile-time error (the struct won't implement `Authorizable`).

### Evolving the Permission System

The permission system is designed to start simple and grow. Here is the progression path:

**Stage 1: Binary access (current)**

System actors are allowed everything. User actors are allowed everything. The infrastructure is in place (actor in context, permission on every request, permission layer wrapping every module) but the `authorize()` method is permissive. This stage validates that the plumbing works without blocking development.

```go
func (p *PermissionLayer) authorize(ctx context.Context, perm Permission) error {
    actor := actor.ActorFrom(ctx)
    if actor == nil {
        return ErrUnauthorized
    }
    return nil // everyone with an actor is allowed
}
```

**Stage 2: Role-based access control (RBAC)**

Introduce roles (e.g., `admin`, `member`, `viewer`) on the `UserActor`. Each role maps to a set of allowed permissions. The `authorize()` method checks whether the actor's role includes the required permission.

```go
// core/actor/roles.go
type Role string

const (
    RoleAdmin  Role = "admin"
    RoleMember Role = "member"
    RoleViewer Role = "viewer"
)

// core/actor/role_permissions.go
var RolePermissions = map[Role][]Permission{
    RoleAdmin:  {"*"},                          // wildcard — all permissions
    RoleMember: {"orders:order:create", "orders:order:view", ...},
    RoleViewer: {"orders:order:view", "inventory:product:view", ...},
}
```

The permission layer checks:

```go
func (p *PermissionLayer) authorize(ctx context.Context, perm Permission) error {
    act := actor.ActorFrom(ctx)
    if act == nil {
        return ErrUnauthorized
    }
    if act.IsSystem() {
        return nil // or check system allowlist (see Stage 4)
    }
    if !act.HasPermission(perm) {
        return ErrForbidden
    }
    return nil
}
```

Where `HasPermission` checks the actor's roles against `RolePermissions`.

**Stage 3: Resource-scoped permissions**

RBAC answers "can this user create orders?" but not "can this user view *this specific* order?" For resource-level access, extend `Authorizable` to optionally provide a resource identifier:

```go
type ResourceAuthorizable interface {
    Authorizable
    ResourceID() string    // e.g., "order:123"
    ResourceOwner() int    // e.g., the user who created the order
}
```

The permission layer can then enforce ownership rules:

```go
// User can view their own orders, admins can view any
func (p *PermissionLayer) authorize(ctx context.Context, req any) error {
    // ... role check first ...
    if ra, ok := req.(ResourceAuthorizable); ok {
        if !act.IsAdmin() && ra.ResourceOwner() != act.UserID() {
            return ErrForbidden
        }
    }
    return nil
}
```

Not every request needs this — only implement `ResourceAuthorizable` on requests that access a specific resource owned by a user.

**Stage 4: System actor allowlists**

Replace the blanket "system actors can do everything" with explicit allowlists per service:

```go
// core/actor/system_permissions.go
var SystemAllowlist = map[string][]Permission{
    "order-worker":    {"orders:order:create", "orders:order:cancel"},
    "billing-worker":  {"orders:order:view", "billing:invoice:create"},
    "migration-tool":  {"*"},
}
```

```go
if act.IsSystem() {
    sysActor := act.(SystemActor)
    if !isAllowed(SystemAllowlist[sysActor.Service], perm) {
        return ErrForbidden
    }
    return nil
}
```

**Stage 5: Policy-based access control (ABAC)**

For complex rules that depend on multiple attributes (time of day, resource state, user department, etc.), replace the role-permission map with a policy engine. The `authorize()` call signature stays the same — only the internals change. Options:

- Custom policy functions per permission
- Integration with an external policy engine (e.g., Open Policy Agent, Casbin)
- Attribute-based rules stored in the database

The key design decision: **the permission layer interface never changes.** Each stage changes what happens inside `authorize()`, but modules continue to declare `Permission()` on their request structs and wrap their API with a permission layer. This means you can evolve access control without touching any business logic.

### Summary of the Evolution

| Stage | What changes | What stays the same |
|-------|-------------|-------------------|
| 1. Binary | `authorize()` allows all | Actor in context, Permission on request, layer wrapping |
| 2. RBAC | Role → permission mapping | Same |
| 3. Resource-scoped | `ResourceAuthorizable` interface on some requests | Same |
| 4. System allowlists | System actor check gets restrictive | Same |
| 5. Policy-based | Internal policy engine | Same |

## Actor System

The actor system identifies who is making a call. It lives in `internal/core/actor/`.

### Actor Types

```go
// core/actor/actor.go
type Actor interface {
    IsSystem() bool
    UserID() int // Returns 0 for system actors
}

type UserActor struct {
    ID    int
    Roles []Role
}

func (a UserActor) IsSystem() bool { return false }
func (a UserActor) UserID() int    { return a.ID }

type SystemActor struct {
    Service string
}

func (a SystemActor) IsSystem() bool { return true }
func (a SystemActor) UserID() int    { return 0 }
```

### Context Propagation

```go
// core/actor/context.go
func WithActor(ctx context.Context, actor Actor) context.Context
func ActorFrom(ctx context.Context) Actor // Returns nil if no actor set
```

### Who Sets the Actor

| Entry point | Actor type | Set by |
|-------------|-----------|--------|
| HTTP handler | `UserActor` | Auth middleware (extracts from JWT) |
| Background worker | `SystemActor` | Worker bootstrap code |
| System-to-system call | `SystemActor` | Calling service |

### Authorization Flow

```
1. Entry point sets Actor on context
2. Permission layer reads Actor from context
3. If SystemActor → allow (Stage 1) or check allowlist (Stage 4+)
4. If UserActor → allow (Stage 1) or check roles (Stage 2+)
5. If no Actor → reject
```

### Rules

- Every call must have an Actor in the context. No Actor = unauthorized.
- The actor package lives in `internal/core/actor/` and is the only cross-cutting dependency modules share (aside from other `core/` packages).

## Core Package

`internal/core/` holds cross-cutting infrastructure that is not a business module. It has no handlers, no services, no API interfaces — just shared plumbing.

```
internal/core/
├── actor/          # Actor system (Actor interface, context helpers)
└── ...             # Future: pagination, error types, middleware helpers, etc.
```

### Rules

- `core/` packages are small and focused. Each solves one cross-cutting concern.
- Modules import from `core/`. `core/` never imports from modules.
- If a `core/` package starts gaining business logic, it probably belongs in a module.

## Provider Pattern

Providers wrap external dependencies (AI clients, email services, payment gateways, cloud storage, etc.) behind interfaces. Business logic depends on the interface, never on the concrete SDK or client.

### Structure

Each provider defines its interface and houses concrete implementations as subpackages:

```
internal/provider/
├── <capability>/
│   ├── provider.go              # Interface definition
│   └── <vendor>/
│       └── client.go            # Concrete implementation
```

Example:

```
internal/provider/
├── ai/
│   ├── provider.go              # AIProvider interface
│   └── anthropic/
│       └── client.go            # Concrete Anthropic implementation
├── email/
│   ├── provider.go              # EmailProvider interface
│   └── sendgrid/
│       └── client.go            # Concrete SendGrid implementation
└── storage/
    ├── provider.go              # StorageProvider interface
    └── s3/
        └── client.go            # Concrete S3 implementation
```

### Interface Definition

```go
// provider/ai/provider.go
package ai

import "context"

type AIProvider interface {
    GenerateText(ctx context.Context, prompt string) (string, error)
    GenerateEmbedding(ctx context.Context, text string) ([]float64, error)
}
```

### Concrete Implementation

```go
// provider/ai/anthropic/client.go
package anthropic

type Client struct {
    apiKey string
    model  string
}

func New(apiKey string) *Client {
    return &Client{apiKey: apiKey, model: "claude-sonnet-4-6"}
}

// implements ai.AIProvider
```

### Wiring

Providers are instantiated in `main.go` and injected into services:

```go
aiProvider := anthropic.New(os.Getenv("AI_API_KEY"))
ordersSvc := orders.NewService(ordersRepo, aiProvider)
```

Swapping a provider (e.g., Anthropic → OpenAI) means writing a new subpackage and changing one line in `main.go`.

### Rules

- The interface lives in the provider's root package (e.g., `provider/ai/provider.go`).
- Concrete implementations live in subpackages named after the vendor (e.g., `provider/ai/anthropic/`).
- Modules depend on the interface, never on the concrete implementation.
- Providers are stateless wrappers — they hold a client/config but no business logic.
- If a provider needs complex configuration, accept a config struct in the constructor rather than many parameters.

## Call Chain

The full call chain for any request:

```
Entry Point (HTTP/Worker/gRPC)
  → Sets Actor on context
    → Permission Layer (checks Actor + Permission)
      → Service (calls req.Validate(), then business logic)
        → Repository Interface (data access)
        │   → Concrete Repository (Jet/PostgreSQL)
        → Provider Interface (external systems)
            → Concrete Provider (vendor SDK)
```

For cross-module calls:

```
ModuleA Service
  → ModuleB PermissionLayer (Actor is already in context)
    → ModuleB Service
      → ModuleB Repository
```

## Wiring

All dependency injection happens in `cmd/webserver/main.go`. Example using japcom modules:

```go
func main() {
    db := connectDB()

    // Repositories
    contentRepo := contentPostgres.New(db)
    learningRepo := learningPostgres.New(db)
    userRepo := userPostgres.New(db)

    // Services wrapped with permission layers
    contentSvc := content.NewService(contentRepo)
    contentAPI := content.NewPermissionLayer(contentSvc)

    learningSvc := learning.NewService(learningRepo, contentAPI)
    learningAPI := learning.NewPermissionLayer(learningSvc)

    userSvc := user.NewService(userRepo, learningAPI, contentAPI)
    userAPI := user.NewPermissionLayer(userSvc)

    // Register HTTP handlers
    router := chi.NewRouter()
    content.RegisterHTTPRoutes(router, contentAPI)
    learning.RegisterHTTPRoutes(router, learningAPI)
    user.RegisterHTTPRoutes(router, userAPI)

    // Start server
    http.ListenAndServe(":8110", router)
}
```

### Rules

- Modules are wired in `main.go`. No module wires itself.
- Services receive their dependencies through constructor injection.
- Permission layers wrap services before being passed to handlers and to other modules.
- The dependency graph must be acyclic.

## Adding a New Module

1. Create the module directory under `internal/modules/<name>/`.
2. Define `api.go` with the public interface and exported types.
3. Define `repository.go` with the repository interface.
4. Define `permissions.go` with permission constants.
5. Implement the service in `service/service.go`.
6. Implement handlers in `handler/http/`.
7. Add the permission layer in `middleware/permission.go`.
8. Add domain types in `domain/`.
9. Implement the repository in `repository/<name>/`.
10. Wire everything in `cmd/webserver/main.go`.

## Adding a New API Method

1. Add the method to the API interface in `api.go`.
2. Create the request struct with `Validate()` and `Permission()` methods.
3. Add the permission constant to `permissions.go`.
4. Implement the method in the service.
5. Add the method to the permission layer.
6. Add the HTTP handler in `handler/http/`.
7. Register the route in `routes.go`.
