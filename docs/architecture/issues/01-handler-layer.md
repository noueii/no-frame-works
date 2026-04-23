# 01 — Handler Layer

> **Severity: HIGH**
> **Axes impacted: Maintainability • Comprehensibility • Reviewability**

## Responsibility

The API handler is the boundary between HTTP and the rest of the system. It has exactly four jobs, and nothing else:

1. **Unwrap** the oapi-generated request object into a service-layer request struct. No manual JSON decoding, no field-by-field validation — those belong elsewhere.
2. **Extract the actor** from `context.Context`. Actor origination is middleware's job; the handler just reads it.
3. **Call the module's API interface** — the service, via an interface type. Never a concrete service struct, never a repository directly.
4. **Map the returned error** to an HTTP status code using `errors.Is` on sentinel errors. Wrap the `*View` result in the oapi response type.

A handler should be 10–20 lines. If it is longer, something it is doing belongs in the service. If it contains business logic, SQL, domain types, or permission rules, it is not a handler anymore.

## What each repo does

### `intranet-cms` — thin, panics on error

`backend/internal/webserver/handler/category.go:33-42`

```go
func (h *Handler) ApiV1CreateCategory(
    ctx context.Context,
    request oapi.ApiV1CreateCategoryRequestObject,
) (oapi.ApiV1CreateCategoryResponseObject, error) {
    result, err := apiservice.CreateCategoryV1(ctx, h.app, request)
    if err != nil {
        panic(fmt.Errorf("[#s1g9gfl2] %w", err))
    }
    return result, nil
}
```

Every handler in `category.go` (11 functions) has this exact shape: delegate to the service, `panic` on error, return the service's oapi response value unchanged. The handler performs no transformation on either side — the service already produces `oapi.*ResponseObject`, so there is nothing for the handler to shape.

### `workflows` — thin, typed-error switch per handler

`backend/internal/webserver/handler/post_api_v1_workflows_new.go:12-65`

```go
func (h *Handler) PostApiV1WorkflowsNew(
    ctx context.Context,
    request oapi.PostApiV1WorkflowsNewRequestObject,
) (oapi.PostApiV1WorkflowsNewResponseObject, error) {
    args := workflow.WorkflowCreateArgs{
        WorkflowTemplateId:        request.Body.WorkflowTemplateId,
        FormFields:                request.Body.FormFields,
        Status:                    oapi.WorkflowStatusEnum(request.Body.Status),
        WorkflowTemplateVersionId: request.Body.WorkflowTemplateVersionId,
    }
    session := h.app.CkidSession().MustGetAuthSessionFromCtx(ctx)

    newWorkflow, err := workflow.CreateWorkflow(
        ctx, h.app, args,
        session.TenantId(), session.OrganizationId(), session.UserId(), &session,
    )
    if err != nil {
        switch e := err.(type) {
        case *errors.ForbiddenError:
            // render + return 403
        case *errors.ValidationError:
            // render + return 400
        default:
            return nil, err
        }
    }
    return oapi.PostApiV1WorkflowsNew200JSONResponse{/* ... */}, nil
}
```

The handler does four things beyond delegation: builds `WorkflowCreateArgs` by copying fields one at a time, pulls the session from the app, threads `tenantId` / `orgId` / `userId` / `session` into the service call as naked parameters, and runs a type switch that maps two specific error types to HTTP responses. The same switch pattern (with different status codes) repeats in every handler file.

### `id-services v3` — thin, session handling copy-pasted

`backend/internal/webserver/v3/handler/teams.go:12-35`

```go
func (h *Handler) CreateTeam(
    ctx context.Context,
    request oapi.CreateTeamRequestObject,
) (oapi.CreateTeamResponseObject, error) {
    sessionData, err := laravel_session.GetSessionData(ctx)
    if err != nil {
        return nil, errors.ErrInvalidSession
    }
    tenant, operator, err := GetTenantAndOperator(ctx, h.app, sessionData)
    if err != nil {
        return nil, err
    }

    result, err := teams.Create(ctx, h.app, &teams.CreateOp{
        Tenant:   *tenant,
        Operator: *operator,
        Name:     request.Body.Name,
    })
    if err != nil {
        return nil, err
    }

    return oapi.CreateTeam201JSONResponse(*result), nil
}
```

The handler does three things before delegating: looks up session data, fetches `tenant` + `operator` via a helper, builds a `CreateOp` struct with those values. Errors pass through unchanged — status mapping happens somewhere else in the stack (presumably in an oapi response writer that inspects typed errors). The 4-line "get session + get tenant+operator" sequence repeats in every handler in `teams.go`, `groups.go`, `users.go`.

## What's good

- **All three are thin.** None decodes HTTP manually, runs business logic, or touches SQL. The handler-service split is structurally present in every codebase.
- **`workflows` and `id-services v3` use typed errors** (rather than string matching) to discriminate outcomes.
- **`id-services v3`** uses per-operation handler methods (`CreateTeam`, `GetTeams`, etc.), not a single dispatcher.
- **All three delegate immediately** to a module-level service function — the handler boundary exists; it is just not doing the work it should be doing.

## What's bad

- **`intranet-cms` panics on every service error.** There is no status discrimination — no 403 for forbidden, no 404 for not found, no 422 for validation. Whatever the service returns becomes a panic, and whatever recovers the panic decides the response. The error tag (`[#s1g9gfl2]`) is the only information the panic carries.
- **Error-to-HTTP mapping is duplicated per handler** in `workflows` and `id-services v3`. A type switch on `*errors.ForbiddenError` / `*errors.ValidationError` lives inside every handler that can produce those errors — identical shape, identical branches, copy-pasted across dozens of files.
- **No actor abstraction on context.** All three thread `session`, `tenantId`, `organizationId`, `userId`, `operator`, `tenant` as explicit parameters from the handler to the service. Nothing on `ctx` answers "who is acting."
- **The service request object is built inline in the handler.** `workflows` builds `WorkflowCreateArgs` by field copy; `id-services v3` builds `CreateOp` with tenant + operator + name. Each is a small copy-paste pattern re-done per handler.
- **Service return type is the oapi response type.** In `intranet-cms` and `id-services v3`, the service returns `oapi.*ResponseObject` or `*oapi.Team` directly. Changing the oapi schema changes the service signature. The handler has no transformation to perform because there is no gap to bridge.

## Improvements — one repo at a time

### `intranet-cms`

1. **Stop panicking.** Replace `panic(fmt.Errorf("[#...] %w", err))` with a sentinel-error → HTTP-status mapping using `errors.Is`. The mapping lives in one helper per module, called by every handler in that module.
2. **Have the service return a module-owned `*View` type**, not `oapi.*ResponseObject`. The handler wraps the `*View` in the oapi response. This breaks the "service signature = HTTP schema" coupling and lets the service evolve independently of the contract.
3. **Move `GetUserWithPermission` to middleware.** Currently the service layer calls it 265 times (see [04-permission-duplication.md](04-permission-duplication.md)); move actor origination into middleware and have the handler read it from `ctx` via `actor.From(ctx)`.

### `workflows`

1. **Replace the per-handler type switch with `errors.Is` on sentinel errors.** One helper function per module handles `ErrForbidden` → 403, `ErrValidation` → 400, `ErrNotFound` → 404. Every handler calls it.
2. **Drop the naked `tenantId`, `organizationId`, `userId`, `session` parameters** from the service call. Attach an `Actor` to `ctx` at middleware time; the handler stops threading them.
3. **Move the `WorkflowCreateArgs` assembly out of the handler.** Define `workflow.CreateWorkflowRequest` in the module root and have the handler do a single field-copy into it — or use a trivial `fromOAPI(body)` helper. The goal is to get the handler down to "delegate + map errors + return."

### `id-services v3`

1. **Move `laravel_session.GetSessionData` + `GetTenantAndOperator` out of every handler.** Per-request tenant/operator resolution is middleware's job. Attach a tenant-scoped actor to `ctx` once per request; every handler stops re-doing the lookup.
2. **Remove `CreateOp` from the handler boundary.** The service should take a request struct with pre-extracted fields (`TenantID`, `Name`); the actor (carrying tenant and operator identity) comes from `ctx`. `CreateOp` currently makes the handler a mini-constructor.
3. **Map errors to HTTP status in the handler** rather than passing through with `return nil, err`. Pulling status decisions back to the boundary where they are read (using sentinels + `errors.Is`) removes the dependency on an out-of-band error translator.
