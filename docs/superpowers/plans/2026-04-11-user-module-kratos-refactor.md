# User Module Kratos Refactor

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix all PR #3 review comments by restructuring the user module to use Kratos identity as the data source and follow the project's oapi-codegen strict server pattern.

**Architecture:** The user repository wraps the identity client (which talks to Kratos) instead of using SQL/go-jet. The handler moves from a standalone chi handler to a method on the webserver's strict server handler. Domain errors move to `domain/errors.go`.

**Tech Stack:** Go, Ory Kratos SDK (`github.com/ory/kratos-client-go`), oapi-codegen strict server, chi router

---

## File Map

| Action | File | Responsibility |
|--------|------|---------------|
| Modify | `backend/kratos/identity.schema.json` | Add `username` trait |
| Modify | `backend/internal/infrastructure/identity/client.go` | Add `GetIdentity`, `UpdateTraits` to interface; add `Username` to `UserDetail` |
| Modify | `backend/internal/infrastructure/identity/kratos.go` | Implement `GetIdentity`, `UpdateTraits`; extract `username` in `GetSession` |
| Modify | `backend/internal/infrastructure/identity/testing.go` | Add stubs for new interface methods; add `Username` to defaults |
| Modify | `backend/internal/modules/user/domain/errors.go` | Add `ErrUserNotFound`, `ErrUsernameTaken` |
| Modify | `backend/internal/modules/user/errors.go` | Remove `ErrUserNotFound`, `ErrUsernameTaken` (keep validation errors) |
| Modify | `backend/internal/modules/user/repository.go` | Change `UpdateUsername` → `Update(ctx, domain.User)` |
| Modify | `backend/repository/user/postgres.go` | Replace `*sql.DB` with `identity.Client` |
| Modify | `backend/repository/user/find_by_id.go` | Rewrite: call `identity.GetIdentity` |
| Modify | `backend/repository/user/find_by_username.go` | Rewrite: call `identity.ListIdentities` + filter by trait |
| Rename+Modify | `backend/repository/user/update_username.go` → `backend/repository/user/update.go` | Rewrite: call `identity.UpdateTraits` with full domain model |
| Modify | `backend/internal/modules/user/service/edit_username/edit_username.go` | Fix fetch-mutate-save; remove sentinel wrapping |
| Delete | `backend/internal/modules/user/handler/http/put_edit_username.go` | Remove raw chi handler |
| Delete | `backend/internal/modules/user/handler/http/handler.go` | Remove standalone handler struct |
| Delete | `backend/internal/modules/user/handler/http/routes.go` | Remove standalone routes |
| Create | `backend/internal/webserver/handler/put_edit_username.go` | Strict server handler for PutEditUsername |
| Create | `backend/internal/webserver/handler/dto_user.go` | `toOAPIUser` conversion |
| Modify | `backend/internal/webserver/handler/handler.go` | Add `userAPI` field + wiring |
| Modify | `backend/internal/webserver/handler/user_handlers.go` | Update `GetUser` stub to use userAPI |

---

### Task 1: Add `username` trait to Kratos identity schema

**Files:**
- Modify: `backend/kratos/identity.schema.json`

- [ ] **Step 1: Update the identity schema**

Add `username` as a string trait alongside `email`. Remove `additionalProperties: false` or keep it — but `username` must be allowed.

```json
{
  "$id": "https://schemas.ory.sh/presets/kratos/identity.email.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "properties": {
        "email": {
          "type": "string",
          "format": "email",
          "title": "E-Mail",
          "ory.sh/kratos": {
            "credentials": {
              "password": {
                "identifier": true
              }
            },
            "verification": {
              "via": "email"
            },
            "recovery": {
              "via": "email"
            }
          }
        },
        "username": {
          "type": "string",
          "title": "Username",
          "minLength": 3,
          "maxLength": 32
        }
      },
      "required": ["email"],
      "additionalProperties": false
    }
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/kratos/identity.schema.json
git commit -m "feat: add username trait to Kratos identity schema"
```

---

### Task 2: Extend identity client interface and implementation

**Files:**
- Modify: `backend/internal/infrastructure/identity/client.go`
- Modify: `backend/internal/infrastructure/identity/kratos.go`
- Modify: `backend/internal/infrastructure/identity/testing.go`

- [ ] **Step 1: Update `client.go` — add `Username` to `UserDetail` and new methods to `Client`**

```go
package identity

import "context"

// Client is a provider-agnostic interface for identity and authentication operations.
// Implementations: KratosClient (production), TestIdentityClient (tests).
type Client interface {
	// Login authenticates with email/password. Returns a session token.
	Login(ctx context.Context, email, password string) (*SessionResult, error)

	// Register creates a new account. Returns a session token.
	Register(ctx context.Context, email, password string) (*SessionResult, error)

	// Logout invalidates the given session token.
	Logout(ctx context.Context, sessionToken string) error

	// GetSession validates a session token and returns the user's identity.
	GetSession(ctx context.Context, sessionToken string) (*UserDetail, error)

	// GetIdentity retrieves an identity by ID.
	GetIdentity(ctx context.Context, id string) (*UserDetail, error)

	// UpdateTraits updates the traits of an identity.
	UpdateTraits(ctx context.Context, id string, traits map[string]interface{}) (*UserDetail, error)

	// ListIdentities returns all identities. Used for trait-based lookups.
	ListIdentities(ctx context.Context) ([]UserDetail, error)
}

// SessionResult is returned after a successful login or registration.
type SessionResult struct {
	SessionToken string
}

// UserDetail contains identity information from the provider.
type UserDetail struct {
	IdentityID string
	Username   string
	Email      string
}
```

- [ ] **Step 2: Implement new methods in `kratos.go`**

Add these methods to `KratosClient`. Also update `GetSession` to extract `username` from traits.

In the existing `GetSession` method, after extracting email, add:

```go
if username, ok := traitsMap["username"].(string); ok {
    detail.Username = username
}
```

Add `GetIdentity`:

```go
func (c *KratosClient) GetIdentity(ctx context.Context, id string) (*UserDetail, error) {
	ident, _, err := c.client.IdentityAPI.GetIdentity(ctx, id).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}

	return c.identityToDetail(ident), nil
}
```

Add `UpdateTraits`:

```go
func (c *KratosClient) UpdateTraits(ctx context.Context, id string, traits map[string]interface{}) (*UserDetail, error) {
	// Fetch current identity to get state and schema_id.
	current, _, err := c.client.IdentityAPI.GetIdentity(ctx, id).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get identity for update: %w", err)
	}

	body := ory.UpdateIdentityBody{
		SchemaId: current.GetSchemaId(),
		State:    string(current.GetState()),
		Traits:   traits,
	}

	updated, _, err := c.client.IdentityAPI.UpdateIdentity(ctx, id).
		UpdateIdentityBody(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update identity traits: %w", err)
	}

	return c.identityToDetail(updated), nil
}
```

Add `ListIdentities`:

```go
func (c *KratosClient) ListIdentities(ctx context.Context) ([]UserDetail, error) {
	identities, _, err := c.client.IdentityAPI.ListIdentities(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list identities: %w", err)
	}

	details := make([]UserDetail, 0, len(identities))
	for _, ident := range identities {
		details = append(details, *c.identityToDetail(&ident))
	}
	return details, nil
}
```

Add shared helper to extract traits from an `*ory.Identity`:

```go
func (c *KratosClient) identityToDetail(ident *ory.Identity) *UserDetail {
	detail := &UserDetail{
		IdentityID: ident.GetId(),
	}

	traits, ok := ident.GetTraitsOk()
	if !ok || traits == nil {
		return detail
	}

	traitsMap, ok := (*traits).(map[string]interface{})
	if !ok {
		return detail
	}

	if email, ok := traitsMap["email"].(string); ok {
		detail.Email = email
	}
	if username, ok := traitsMap["username"].(string); ok {
		detail.Username = username
	}

	return detail
}
```

Refactor `GetSession` to use `identityToDetail`:

```go
func (c *KratosClient) GetSession(ctx context.Context, sessionToken string) (*UserDetail, error) {
	session, _, err := c.client.FrontendAPI.ToSession(ctx).
		XSessionToken(sessionToken).Execute()
	if err != nil {
		return nil, fmt.Errorf("kratos session check failed: %w", err)
	}

	identity := session.GetIdentity()
	return c.identityToDetail(&identity), nil
}
```

- [ ] **Step 3: Update `testing.go` — add stubs for new interface methods**

```go
package identity

import "context"

// TestIdentityClient is a configurable test double for identity.Client.
type TestIdentityClient struct {
	ResSession    *SessionResult
	ResMeDetail   *UserDetail
	ResIdentities []UserDetail
	Err           error
}

func (c *TestIdentityClient) Login(_ context.Context, _, _ string) (*SessionResult, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	return c.ResSession, nil
}

func (c *TestIdentityClient) Register(_ context.Context, _, _ string) (*SessionResult, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	return c.ResSession, nil
}

func (c *TestIdentityClient) Logout(_ context.Context, _ string) error {
	return c.Err
}

func (c *TestIdentityClient) GetSession(_ context.Context, _ string) (*UserDetail, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	return c.ResMeDetail, nil
}

func (c *TestIdentityClient) GetIdentity(_ context.Context, _ string) (*UserDetail, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	return c.ResMeDetail, nil
}

func (c *TestIdentityClient) UpdateTraits(_ context.Context, _ string, _ map[string]interface{}) (*UserDetail, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	return c.ResMeDetail, nil
}

func (c *TestIdentityClient) ListIdentities(_ context.Context) ([]UserDetail, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	return c.ResIdentities, nil
}

// GetDefaultTestIdentityClient returns a TestIdentityClient with sensible defaults.
func GetDefaultTestIdentityClient() *TestIdentityClient {
	return &TestIdentityClient{
		ResSession: &SessionResult{SessionToken: "test-session-token"},
		ResMeDetail: &UserDetail{
			IdentityID: "00000000-0000-0000-0000-000000000001",
			Username:   "testuser",
			Email:      "test@example.com",
		},
	}
}
```

- [ ] **Step 4: Verify it compiles**

Run: `cd backend && go build ./internal/infrastructure/identity/...`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add backend/internal/infrastructure/identity/
git commit -m "feat: extend identity client with GetIdentity, UpdateTraits, ListIdentities"
```

---

### Task 3: Move domain errors

**Files:**
- Modify: `backend/internal/modules/user/domain/errors.go`
- Modify: `backend/internal/modules/user/errors.go`

- [ ] **Step 1: Add domain errors to `domain/errors.go`**

```go
package domain

import "errors"

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUsernameTaken = errors.New("username is already taken")
)
```

- [ ] **Step 2: Remove domain errors from `errors.go`**

Keep only validation errors:

```go
package user

import "errors"

var (
	ErrUserIDRequired   = errors.New("user_id is required")
	ErrUsernameRequired = errors.New("username is required")
	ErrUsernameTooShort = errors.New("username must be at least 3 characters")
	ErrUsernameTooLong  = errors.New("username must be at most 32 characters")
)
```

- [ ] **Step 3: Update all references from `user.ErrUserNotFound` → `domain.ErrUserNotFound`**

Files that reference these errors:
- `backend/internal/modules/user/service/edit_username/edit_username.go` — uses `user.ErrUserNotFound`, `user.ErrUsernameTaken`

Update imports:
```go
// change:
"github.com/noueii/no-frame-works/internal/modules/user"
// to include:
"github.com/noueii/no-frame-works/internal/modules/user/domain"
```

Replace:
- `user.ErrUserNotFound` → `domain.ErrUserNotFound`
- `user.ErrUsernameTaken` → `domain.ErrUsernameTaken`

- [ ] **Step 4: Verify it compiles**

Run: `cd backend && go build ./internal/modules/user/...`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add backend/internal/modules/user/domain/errors.go backend/internal/modules/user/errors.go backend/internal/modules/user/service/
git commit -m "refactor: move domain errors to domain/errors.go"
```

---

### Task 4: Update repository interface and rewrite implementation

**Files:**
- Modify: `backend/internal/modules/user/repository.go`
- Modify: `backend/repository/user/postgres.go`
- Modify: `backend/repository/user/find_by_id.go`
- Modify: `backend/repository/user/find_by_username.go`
- Delete: `backend/repository/user/update_username.go`
- Create: `backend/repository/user/update.go`

- [ ] **Step 1: Update the repository interface**

`backend/internal/modules/user/repository.go`:

```go
package user

import (
	"context"

	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

// UserRepository defines the data access contract for the user module.
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, u domain.User) (*domain.User, error)
}
```

- [ ] **Step 2: Rewrite `postgres.go` to wrap identity client**

Rename the struct since it's no longer Postgres-specific. The file stays at the same path for minimal churn.

```go
package user

import (
	"github.com/noueii/no-frame-works/internal/infrastructure/identity"
	usermod "github.com/noueii/no-frame-works/internal/modules/user"
)

// Repository implements user.UserRepository using the identity provider.
type Repository struct {
	identity identity.Client
}

// New creates a new user repository backed by the identity provider.
func New(identity identity.Client) usermod.UserRepository {
	return &Repository{identity: identity}
}
```

- [ ] **Step 3: Rewrite `find_by_id.go`**

```go
package user

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

func (r *Repository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	detail, err := r.identity.GetIdentity(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get identity: %w", err)
	}

	return toDomain(detail), nil
}
```

- [ ] **Step 4: Rewrite `find_by_username.go`**

Uses `ListIdentities` and filters by username trait. Returns `nil, nil` if not found (matching the existing contract).

```go
package user

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

func (r *Repository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	identities, err := r.identity.ListIdentities(ctx)
	if err != nil {
		return nil, fmt.Errorf("list identities: %w", err)
	}

	for _, detail := range identities {
		if detail.Username == username {
			return toDomain(&detail), nil
		}
	}

	return nil, nil
}
```

- [ ] **Step 5: Delete `update_username.go` and create `update.go`**

Delete `backend/repository/user/update_username.go`.

Create `backend/repository/user/update.go`:

```go
package user

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

func (r *Repository) Update(ctx context.Context, u domain.User) (*domain.User, error) {
	traits := map[string]interface{}{
		"email":    u.Email,
		"username": u.Username,
	}

	detail, err := r.identity.UpdateTraits(ctx, u.ID, traits)
	if err != nil {
		return nil, fmt.Errorf("update identity traits: %w", err)
	}

	return toDomain(detail), nil
}
```

- [ ] **Step 6: Add `toDomain` helper**

Create `backend/repository/user/mapping.go`:

```go
package user

import (
	"github.com/noueii/no-frame-works/internal/infrastructure/identity"
	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

func toDomain(detail *identity.UserDetail) *domain.User {
	return &domain.User{
		ID:       detail.IdentityID,
		Username: detail.Username,
		Email:    detail.Email,
	}
}
```

- [ ] **Step 7: Verify it compiles**

Run: `cd backend && go build ./repository/user/...`
Expected: no errors

- [ ] **Step 8: Commit**

```bash
git add backend/internal/modules/user/repository.go backend/repository/user/
git rm backend/repository/user/update_username.go
git commit -m "refactor: rewrite user repository to wrap identity client"
```

---

### Task 5: Fix the service layer

**Files:**
- Modify: `backend/internal/modules/user/service/edit_username/edit_username.go`

- [ ] **Step 1: Rewrite `edit_username.go` with fetch-mutate-save and fix error wrapping**

```go
package editusername

import (
	"context"
	"fmt"

	"github.com/noueii/no-frame-works/internal/modules/user"
	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

// Execute changes a user's username.
func Execute(
	ctx context.Context,
	repo user.UserRepository,
	req user.EditUsernameRequest,
) (user.UserView, error) {
	if err := req.Validate(); err != nil {
		return user.UserView{}, err
	}

	// Fetch existing user.
	existing, err := repo.FindByID(ctx, req.UserID)
	if err != nil {
		return user.UserView{}, fmt.Errorf("failed to find user: %w", err)
	}
	if existing == nil {
		return user.UserView{}, domain.ErrUserNotFound
	}

	// Check username uniqueness.
	taken, err := repo.FindByUsername(ctx, req.Username)
	if err != nil {
		return user.UserView{}, fmt.Errorf("failed to check username: %w", err)
	}
	if taken != nil && taken.ID != req.UserID {
		return user.UserView{}, domain.ErrUsernameTaken
	}

	// Mutate and save.
	existing.Username = req.Username
	updated, err := repo.Update(ctx, *existing)
	if err != nil {
		return user.UserView{}, fmt.Errorf("failed to update user: %w", err)
	}

	return user.UserView{
		ID:       updated.ID,
		Username: updated.Username,
		Email:    updated.Email,
	}, nil
}
```

Key changes from the original:
- `req.Validate()` error returned directly (no `fmt.Errorf("validation failed: %w", err)` wrapping)
- `user.ErrUserNotFound` → `domain.ErrUserNotFound`
- `user.ErrUsernameTaken` → `domain.ErrUsernameTaken`
- `repo.UpdateUsername(ctx, req.UserID, req.Username)` → fetch-mutate-save: `existing.Username = req.Username` then `repo.Update(ctx, *existing)`

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./internal/modules/user/...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/modules/user/service/edit_username/edit_username.go
git commit -m "fix: service uses fetch-mutate-save, returns sentinel errors directly"
```

---

### Task 6: Delete standalone handler and add strict server handler

**Files:**
- Delete: `backend/internal/modules/user/handler/http/put_edit_username.go`
- Delete: `backend/internal/modules/user/handler/http/handler.go`
- Delete: `backend/internal/modules/user/handler/http/routes.go`
- Create: `backend/internal/webserver/handler/put_edit_username.go`
- Create: `backend/internal/webserver/handler/dto_user.go`
- Modify: `backend/internal/webserver/handler/handler.go`
- Modify: `backend/internal/webserver/handler/user_handlers.go`

- [ ] **Step 1: Delete the standalone handler files**

```bash
rm backend/internal/modules/user/handler/http/put_edit_username.go
rm backend/internal/modules/user/handler/http/handler.go
rm backend/internal/modules/user/handler/http/routes.go
```

If this leaves an empty `handler/http/` directory, remove it too:
```bash
rmdir backend/internal/modules/user/handler/http/ 2>/dev/null
rmdir backend/internal/modules/user/handler/ 2>/dev/null
```

- [ ] **Step 2: Create `dto_user.go`**

`backend/internal/webserver/handler/dto_user.go`:

```go
package handler

import (
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/modules/user"
)

func toOAPIUser(v user.UserView) oapi.User {
	return oapi.User{
		Id:       uuid.MustParse(v.ID),
		Username: v.Username,
		Name:     v.Username,
		Email:    openapi_types.Email(v.Email),
	}
}
```

Note: `Name` is set to `Username` since the shared schema requires both fields. Adjust if `Name` has a different source.

- [ ] **Step 3: Create `put_edit_username.go`**

`backend/internal/webserver/handler/put_edit_username.go`:

```go
package handler

import (
	"context"
	"errors"

	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/modules/user"
	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

// PutEditUsername handles PUT /users/{id}/username.
func (h *Handler) PutEditUsername(ctx context.Context, request oapi.PutEditUsernameRequestObject) (oapi.PutEditUsernameResponseObject, error) {
	result, err := h.userAPI.EditUsername(ctx, user.EditUsernameRequest{
		UserID:   request.Id.String(),
		Username: request.Body.Username,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return oapi.PutEditUsername404JSONResponse{Error404JSONResponse: oapi.Error404JSONResponse{
				ErrorCode:    "not_found",
				ErrorMessage: "user not found",
			}}, nil
		}
		if errors.Is(err, domain.ErrUsernameTaken) {
			return oapi.PutEditUsername409JSONResponse{Error409JSONResponse: oapi.Error409JSONResponse{
				ErrorCode:    "conflict",
				ErrorMessage: "username is already taken",
			}}, nil
		}
		return oapi.PutEditUsername400JSONResponse{Error400JSONResponse: oapi.Error400JSONResponse{
			ErrorCode:    "bad_request",
			ErrorMessage: err.Error(),
		}}, nil
	}

	return oapi.PutEditUsername200JSONResponse(toOAPIUser(result)), nil
}
```

- [ ] **Step 4: Wire up user module in `handler.go`**

Update `backend/internal/webserver/handler/handler.go` to add the user API:

```go
package handler

import (
	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/infrastructure/identity"
	"github.com/noueii/no-frame-works/internal/modules/post"
	postmw "github.com/noueii/no-frame-works/internal/modules/post/middleware"
	postservice "github.com/noueii/no-frame-works/internal/modules/post/service"
	postrepo "github.com/noueii/no-frame-works/repository/post"
	"github.com/noueii/no-frame-works/internal/modules/user"
	usermw "github.com/noueii/no-frame-works/internal/modules/user/middleware"
	userservice "github.com/noueii/no-frame-works/internal/modules/user/service"
	userrepo "github.com/noueii/no-frame-works/repository/user"
)

type Handler struct {
	oapi.StrictServerInterface

	app      *config.App
	identity identity.Client
	postAPI  post.PostAPI
	userAPI  user.UserAPI
}

func NewHandler(app *config.App) *Handler {
	repo := postrepo.New(app.DB())
	svc := postservice.New(repo)
	api := postmw.NewPermissionLayer(svc, repo)

	idClient := app.IdentityClient()
	userRepo := userrepo.New(idClient)
	userSvc := userservice.New(userRepo)
	userAPI := usermw.NewPermissionLayer(userSvc)

	return &Handler{
		app:      app,
		identity: idClient,
		postAPI:  api,
		userAPI:  userAPI,
	}
}
```

- [ ] **Step 5: Update `user_handlers.go` — remove stub GetUser**

Replace the stub with a real implementation or keep it as a stub that acknowledges the user module exists. For now, keep it minimal but use the user API if a `GetUser` method exists on the API interface. Since `UserAPI` only has `EditUsername`, keep the stub:

`backend/internal/webserver/handler/user_handlers.go`:

```go
package handler

import (
	"context"

	"github.com/noueii/no-frame-works/generated/oapi"
)

// GetUser handles GET /users/{id}. Stub — returns 404 for now.
func (h *Handler) GetUser(_ context.Context, _ oapi.GetUserRequestObject) (oapi.GetUserResponseObject, error) {
	return oapi.GetUser404JSONResponse{Error404JSONResponse: oapi.Error404JSONResponse{
		ErrorCode:    "not_found",
		ErrorMessage: "user not found",
	}}, nil
}
```

- [ ] **Step 6: Verify it compiles**

Run: `cd backend && go build ./...`
Expected: no errors

- [ ] **Step 7: Commit**

```bash
git rm backend/internal/modules/user/handler/http/put_edit_username.go backend/internal/modules/user/handler/http/handler.go backend/internal/modules/user/handler/http/routes.go
git add backend/internal/webserver/handler/
git commit -m "refactor: move user handler to strict server pattern"
```

---

### Task 7: Verify full build and cleanup

- [ ] **Step 1: Build the entire project**

Run: `cd backend && go build ./...`
Expected: no errors

- [ ] **Step 2: Run vet**

Run: `cd backend && go vet ./...`
Expected: no issues

- [ ] **Step 3: Check for unused imports or references to deleted types**

Grep for any remaining references to the old patterns:

```bash
grep -r "UpdateUsername" backend/ --include="*.go" | grep -v "_test.go" | grep -v "generated/"
grep -r "user\.ErrUserNotFound\|user\.ErrUsernameTaken" backend/ --include="*.go"
grep -r "modules/user/handler/http" backend/ --include="*.go"
```

All three should return no results.

- [ ] **Step 4: Commit any cleanup**

If step 3 found issues, fix and commit:

```bash
git add -A
git commit -m "chore: cleanup stale references"
```
