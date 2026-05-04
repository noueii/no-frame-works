# 02 — Application Layer (Service)

> **Severity: HIGH**
> **Axes impacted: Maintainability • Comprehensibility • Reviewability**

## Responsibility

The application layer — the service — is where the **business operation** lives. It is the only layer that orchestrates the steps of a feature end-to-end, composing primitives from the layers above and below.

The service function is a **thin, linear script**:

1. `req.Validate()` — input-shape validation, declared on the request struct.
2. `req.CheckPermission(ctx[, model])` — authorization, declared on the same struct.
3. For updates: fetch the existing domain model via the repository, mutate it in memory, save it back. For creates: build a new domain model via a factory, save it.
4. Return `*View` on success, `nil, err` on failure.

The service owns the **orchestration**, not the work. It does not construct `model.*` structs, it does not run SQL, it does not decide HTTP status codes, and it does not re-implement permission rules. If a concern is not orchestration, it belongs in a different layer.

A service function should be 15–40 lines for most operations. Longer means it is absorbing concerns that should live elsewhere.

## What each repo does

### `intranet-cms` — 127-line god function

`backend/internal/app/appservice/apiservice/category.go:24-150`. `CreateCategoryV1` does, in order: extract the DB handle, call permission check inline (7 lines), null-check the request body, run struct-tag validation, begin a SQL transaction with `defer tx.Rollback()`, loop over the request to collect user-group IDs, query the repository for user groups, loop over the request to collect user IDs, query for org users, query the category count per org, enforce a tenant-limit business rule (50 categories per org), build a `model.CategoryTbl` from the request, insert the category row, build and insert `UserGroupCategoryTbl` rows, build and insert `UserCategoryTbl` rows, commit, return a 200 response. Twelve responsibilities, no internal boundaries, 127 lines.

Condensed:

```go
func CreateCategoryV1(
    ctx context.Context,
    app *config.App,
    request oapi.ApiV1CreateCategoryRequestObject,
) (oapi.ApiV1CreateCategoryResponseObject, error) {
    db := app.DB()
    authUser, err := useauth.GetUserWithPermission(ctx, db, permission.IsCKIDAdmin)
    // 7 lines of permission check + error mapping

    if request.Body == nil { return &oapi.ApiV1CreateCategory422Response{}, nil }
    err = app.Validator().StructCtx(ctx, request.Body)
    // 4 lines of validation

    tx, err := db.BeginTx(ctx, nil)
    defer tx.Rollback()
    // ... 90 lines of FK lookups, mapping, business rules, three INSERT calls ...
    tx.Commit()
    return oapi.ApiV1CreateCategory200Response{}, nil
}
```

Every sibling function in the same file (`GetCategoriesV1`, `UpdateCategoryV1`, `DeleteCategoryV1`, `MergeCategoriesV1`, ...) has the same shape. The pattern is the convention, not an exception.

### `workflows` — 120 lines with seven private helpers

`backend/internal/app/app_service/workflow/create_workflow.go:40-159`. `CreateWorkflow` delegates to seven private helpers, all defined in the same package:

```go
func CreateWorkflow(
    ctx context.Context,
    app *config.App,
    args WorkflowCreateArgs,
    tenantId, organizationId, userId uuid.UUID,
    session *session.AuthenticatedSession,
) (*WorkflowCreateResult, error) {
    tpl, err := validateWorkflowTemplate(ctx, app, tenantId, organizationId, args.WorkflowTemplateId, args.WorkflowTemplateVersionId)
    err = validateFormFieldsAndPermissions(ctx, app, session, tenantId, organizationId, userId, args.FormFields, tpl)

    err = app.DB().WithTransaction(nil, func(rootTx *sql.Tx) error {
        workflowWithEvents, err := processWorkflowCreation(ctx, app, rootTx, userId, args.FormFields, args.Status, tpl)
        row, err := insertWorkflowAndHandleFiles(ctx, app, rootTx, tenantId, organizationId, args.FormFields, workflowWithEvents)
        err = insertStepFulls(ctx, rootTx, workflowWithEvents.OverriddenApprovalSteps())
        err = insertStepFulls(ctx, rootTx, workflowWithEvents.SelectedApproverSteps())
        err = mutation.IncrementWorkflowTemplateUsageCount(ctx, rootTx, workflowWithEvents.Workflow.WorkflowTemplateID)
        workflow, workflowVersion = buildWorkflowResult(row)
        return nil
    })

    err = handlePostTransactionTasks(ctx, app, tenantId, organizationId, workflow)
    return &WorkflowCreateResult{Workflow: workflow, WorkflowVersion: workflowVersion}, nil
}
```

The helpers (`validateWorkflowTemplate`, `validateFormFieldsAndPermissions`, `processWorkflowCreation`, `insertWorkflowAndHandleFiles`, `insertStepFulls`, `IncrementWorkflowTemplateUsageCount`, `buildWorkflowResult`, `handlePostTransactionTasks`) are private to the `app_service/workflow` package. They all take `*config.App`, naked `uuid.UUID` identity parameters, and internal DTOs. The shape looks like orchestration; the substance is one god function broken into named pieces with no contracts between them.

### `id-services v3` — `CreateOp` struct with four methods

`backend/internal/app/v3/service/teams/create.go:19-100`. `Create` is a 10-line orchestration that calls four methods on a per-operation struct:

```go
type CreateOp struct {
    Tenant   model.Tenant
    Operator dto.UserWithOrgAndGroupInfo
    Name     string

    organization *model.Organization // mutable state populated by performMutation
}

func Create(ctx context.Context, app *config.App, op *CreateOp) (*oapi.Team, error) {
    if err := op.checkPermission(); err != nil { return nil, err }
    if err := op.checkNameUniqueness(ctx, app); err != nil { return nil, err }
    if err := op.performMutation(ctx, app); err != nil { return nil, err }
    return op.buildResult(), nil
}
```

Each method is 10–15 lines. `checkPermission` calls `permission.IsTenantAdmin(&op.Operator)` and returns `*errors.HttpError{Code: http.StatusForbidden}` on failure. `checkNameUniqueness` calls `organization_repo.GetByTenantAndName(ctx, app.DB(), ...)` directly and returns `*errors.HttpError{Code: http.StatusConflict}` on duplicate. `performMutation` constructs `model.Organization{}` inline with `uuid.New()` and calls `mutation.InsertOrganization(ctx, app.DB(), ...)`. `buildResult` returns `*oapi.Team`.

This is genuinely the cleanest of the three. Operations live one-per-file with unit tests (`create_test.go`). The `Create` body is a readable orchestration. I want to be fair about that — v3 is what "clean up the service without changing the architecture" looks like when it is done well.

## What's good

- **`id-services v3` has one file per operation with tests alongside.** `create.go`, `update.go`, `archive.go`, each with matching `_test.go`. Closest layout to what `no-frame-works` recommends.
- **`id-services v3` decomposes the operation into named steps** that match a clear mental model (permission → uniqueness → mutation → result). A reader understands the shape in ten seconds.
- **`workflows` uses a `db.WithTransaction(fn)` helper** rather than inline `BeginTx`/`Commit`/`Rollback` — closer to the `TxManager` pattern the `no-frame-works` template proposes.
- **`id-services v3`'s `checkPermission` is a method**, not inline code. One step away from being on a request struct, which is where the no-frame-works pattern puts it.
- **All three use `github.com/google/uuid`** for ID generation consistently — not scattered between stdlib, DB auto-gen, and third-party libs.

## What's bad

- **Every service function mixes orchestration with infrastructure.** All three call directly into `repository.*`, `mutation.*`, or `apiservice`-internal query helpers. There is no repository **interface** that abstracts the data source.
- **Every service function constructs `model.*` structs directly.** No domain layer: `model.CategoryTbl`, `model.Organization`, `internal_dto.Workflow` are built inline with field literals. No factory, no invariants, no shared construction logic.
- **Permission, validation, and business logic are inline at the entry of the function.** Nothing prevents a new service from omitting a validation step or a permission check. The ritual is remembered by the author, not enforced by the type system.
- **Transaction management is ad-hoc across all three.** `intranet-cms` does inline `BeginTx` + `defer Rollback` per service. `workflows` uses a per-service `WithTransaction(fn)` helper but declares transactions per-function. `id-services v3` does not wrap `Create` in a transaction at all — `performMutation` runs exactly one INSERT, which works only because the operation happens to be single-write.
- **Return types cross layer boundaries.** `intranet-cms` returns `oapi.*ResponseObject`; `workflows` returns `*WorkflowCreateResult` built from internal DTOs that alias oapi types; `id-services v3` returns `*oapi.Team`. There is no module-owned `View` type that the service can evolve independently of the HTTP schema.
- **`id-services v3`'s `CreateOp` is input + mutable state in one struct.** The `organization` field is written by `performMutation` and read by `buildResult`. The struct doubles as a parameter bag and a state machine, which means no method has a pure, testable signature — each implicitly depends on the order of prior method calls.

## Improvements — one repo at a time

### `intranet-cms`

1. **Break the god function into four layers before touching anything else.** Handler → service (with a request struct that owns `Validate` + `CheckPermission`) → domain (with `NewCategory` factory, `MaxCategoriesPerOrg` constant, and `ErrCategoryLimitReached` sentinel) → repository (with an interface). This is the single highest-leverage change in the codebase.
2. **Define a `Service` struct per module** implementing a `category.API` interface. Stop using package-level functions. The service's dependencies (`repo`, `tx`, `notifier`) are injected via the constructor.
3. **Introduce `domain.Category`** separate from `model.CategoryTbl`. The service passes `domain.*` to the repository; only the repository knows about `model.*`.
4. **Move timestamp and invariant enforcement into the domain factory.** `domain.NewCategory(...)` sets the ID, timestamps, and default flags; the service stops building these inline.

### `workflows`

1. **Collapse the seven private helpers into layer-appropriate homes.** Validation helpers → `req.Validate()`. Permission helpers → `req.CheckPermission(ctx, tpl)`. Mutation helpers → repository methods that own the full write. Post-transaction helpers → an injected side-effect collaborator (`notifier`, `eventBus`).
2. **Replace the private-helper-as-seam pattern with interfaces.** Every step currently calling `helperFn(ctx, app, ...)` should call `s.something.Method(ctx, ...)` where `something` is an injected interface. The decomposition becomes real rather than cosmetic.
3. **Extract domain logic into methods on a `domain.Workflow` type.** `processWorkflowCreation` currently generates approval steps, evaluates form fields, and builds the internal state. That is business logic with no clear type home — it belongs on `domain.Workflow` with methods like `GenerateApprovalSteps()` and `EvaluateFormFields()`.
4. **Replace `*WorkflowCreateResult` with a module-owned `workflow.View`.** The current result struct is a bag of oapi-aliased DTOs; the service boundary should use a type owned by the `workflow` module.

### `id-services v3`

1. **Split `CreateOp` into a read-only request struct and local variables.** `Tenant`, `Operator`, `Name` become fields on `team.CreateRequest`; `organization` becomes a local variable in the service method. The struct stops doubling as a state machine, and every method's contract becomes explicit.
2. **Replace direct `organization_repo.*` and `mutation.*` calls with a `team.Repository` interface** owned by the module. Inject the implementation through the service constructor. Fake it in tests.
3. **Introduce `domain.Team` with a `NewTeam(...)` factory.** `performMutation` stops building `model.Organization` inline; the service calls `domain.NewTeam(...)` and passes it to `s.repo.Create(ctx, t)`.
4. **Replace `*errors.HttpError` returns with sentinel errors** (`team.ErrNameAlreadyTaken`, `team.ErrForbidden`) declared in `team/errors.go`. The handler then maps sentinels to status codes via `errors.Is` — the service stops knowing about HTTP.
5. **Adopt `TxManager.WithTransaction`.** Right now `Create` has no transaction wrapping because it happens to do one write. The moment a second write is added (e.g. creating a default group when a team is created), you need atomicity — and the current shape has no clean way to declare it.
