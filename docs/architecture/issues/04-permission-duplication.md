# 04 — Permission checks duplicated inline at every call site

> **Severity: CRITICAL**
> **Axes impacted: Maintainability • Comprehensibility • Reviewability**

## Summary

In all three reference codebases, the permission check at the start of a service function is a block of code that lives **at the call site**, not behind a boundary that enforces it. The shape varies — `intranet-cms` copy-pastes a 7-line block verbatim, `workflows` hides it behind helper functions, `id-services` has a typed permission package — but in every case **the dispatch is inline, positional, and optional**. Nothing structural prevents a new service function from being written without a permission check, and nothing surfaces a missing check in code review.

In `intranet-cms` alone, the permission block appears **265 times across 76 files**. The error-handling branch that maps it to a 403 response appears **258 times across 73 files**. Those numbers are the cost of "inline permission" at scale: every change to how authorization is checked or reported is a hundreds-of-files PR.

## What you have to know

To change **one thing** about authorization — a permission rule, an error response, a log line, the shape of the `Forbidden` error — you have to know:

1. **Every call site** of the permission helper, because each one is an independent decision point.
2. **Every function signature** that threads identity parameters, because they're naked `uuid.UUID`s or `*session.AuthenticatedSession`, not a context value.
3. **Every error path**, because the inline block also duplicates error wrapping and HTTP response mapping.

To **add a new endpoint**, you have to remember to:

1. Copy the permission block in.
2. Pick the right permission constant.
3. Translate the error to a 403 (not a 500).
4. Not forget any of those three.

To **review an endpoint** for correct authorization, you have to read the whole service function, find the permission block (if any), verify it checks the right permission for this operation, verify the error maps to 403 not 500, and compare against the other 264 sites to confirm the pattern matches.

The core problem: **the absence of a permission check is visually identical to its presence when you're scrolling a diff**. The compiler cannot help. The linter cannot help. The reviewer has to remember.

## Investigation

### intranet-cms — 265 inline copies across 76 files

The block, repeated verbatim across the codebase with only the permission constant and the error code changing:

```go
// backend/internal/app/appservice/apiservice/category.go:31-37
authUser, err := useauth.GetUserWithPermission(ctx, db, permission.IsCKIDAdmin)
if err != nil {
    if errors.Is(err, useauth.ErrPermission) {
        return &oapi.ApiV1CreateCategory403Response{}, nil
    }
    return nil, fmt.Errorf("[#tg2hz3tc] %w", err)
}
orgUser := authUser.OrgUser
```

And again, in a completely different feature:

```go
// backend/internal/app/appservice/apiservice/link.go:110-116
authUser, err := useauth.GetUserWithPermission(ctx, db, permission.CanEditLayout)
if err != nil {
    if errors.Is(err, useauth.ErrPermission) {
        return &oapi.ApiV1CreateLink403Response{}, nil
    }
    return nil, fmt.Errorf("[#a6nba9u5] %w", err)
}
orgUser := authUser.OrgUser
```

Same shape, same three branches, same error wrapping. The only differences per site are (1) the permission constant and (2) the generated oapi 403 response type.

**Measured**:
```
rg -c "GetUserWithPermission" backend/          → 265 across 76 files
rg -c "errors.Is(err, useauth.ErrPermission)"   → 258 across 73 files
```

Every new endpoint is one more copy. Every change to how permission failures are reported — say, adding a log line, or renaming the error — is a change that touches 258 files. Adding a second kind of permission failure (say, a `RateLimitExceeded` branch) means editing each of those 258 sites to decide which of two branches to take.

And the 7 lines is the *minimum* version. Heavier endpoints interleave more logic into the same block (org lookup, feature flag checks, per-tenant overrides) and the copy-paste drifts. There is no "the" version of the block anymore — there are 265 dialects.

### workflows — extracted helpers, still inline dispatch

`workflows` factored authorization into a dedicated `permission_service` package. Permission logic is no longer in the service file itself — but the **dispatch** still is, and the service signature still threads every identity parameter by hand:

```go
// backend/internal/app/app_service/workflow/create_workflow.go:40-78
func CreateWorkflow(
    ctx context.Context,
    app *config.App,
    args WorkflowCreateArgs,
    tenantId uuid.UUID,
    organizationId uuid.UUID,
    userId uuid.UUID,
    session *session.AuthenticatedSession,
) (*WorkflowCreateResult, error) {
    workflowTemplateWithConditions, err := validateWorkflowTemplate(
        ctx, app, tenantId, organizationId,
        args.WorkflowTemplateId, args.WorkflowTemplateVersionId,
    )
    if err != nil { return nil, err }

    err = validateFormFieldsAndPermissions(
        ctx, app, session,
        tenantId, organizationId, userId,
        args.FormFields, workflowTemplateWithConditions,
    )
    if err != nil { return nil, err }
    // ...
}
```

What improved: the permission logic moved out of the service function. What didn't:

- Every service function still takes `tenantId`, `organizationId`, `userId`, `session` as separate parameters and has to thread them through.
- The call to `validateFormFieldsAndPermissions` is still inline. Nothing prevents a new service function from skipping it.
- The same pattern repeats: `edit_workflow.go:134-184` (`retrieveAndValidateWorkflow`) has a hand-written "user is submitter" check at line 164, and **the same check appears verbatim in `edit_draft_workflow.go:63`**. A helper got extracted once and then copy-pasted again.

This is the "halfway" state: the **code** for the permission check is centralized, but the **decision to check permission** is still a per-site, remember-or-forget event.

### id-services — typed errors, still inline dispatch

`id-services` went further. Permission logic lives in its own `shared/permission/` package, and failures are typed `ForbiddenError`s:

```go
// backend/internal/app/v1/shared/permission/team_user_group.go:14-62
func CanManageTeamUserGroup(
    ctx context.Context,
    app *config.App,
    tenantId uuid.UUID,
    orgId uuid.UUID,
    userId uuid.UUID,
) error {
    if isTenantAdmin(ctx, app, tenantId, userId) {
        return nil
    }
    orgMember, err := organization_member_repo.GetByOrgAndUser(
        ctx, app.DB(), tenantId, orgId, userId,
    )
    if err != nil {
        return utils.WrapErr(err)
    }
    if orgMember == nil {
        return utils.WrapErr(&errors.ForbiddenError{...})
    }
    if orgMember.Role == int32(dto.UserOrgRoleAdmin) {
        return nil
    }
    return utils.WrapErr(&errors.ForbiddenError{...})
}
```

Invoked at the entry of every service:

```go
// backend/internal/app/v1/service/team_user_groups/create.go:25-34
err := permission.CanManageTeamUserGroup(
    ctx, app,
    payload.TenantId,
    payload.OrganizationId,
    payload.OperatorId,
)
if err != nil {
    return nil, utils.WrapErr(err)
}
```

This is the strongest of the three — centralized logic, typed errors, no copy-pasted error mapping. But it has the **same structural weakness**: the call is inline and positional, each service function has to remember to make it, and the permission helper's signature is naked `uuid.UUID`s rather than something tied to the request.

Forgetting the call is a silent 200 OK for an unauthorized request. Adding a new dimension to the permission check (e.g. "check against the target model, not just the actor") means changing the helper signature and updating every call site.

## Impact on the three axes

**Maintainability.**
Changing the permission system in `intranet-cms` is a 258-file PR. Even in `id-services`, where the logic is centralized, any change that affects the *shape* of the permission call — a new parameter, a new error type, a new branch — ripples across every service that invokes it. The inline dispatch is what makes the cost linear in the number of endpoints instead of constant.

**Comprehensibility.**
You cannot answer "what can a user do in this system" without reading every service function. There is no module-level summary of authorization. A new engineer opening the `post` module should be able to see, in one file, every permission rule the module enforces. In all three reference codebases, they cannot.

**Reviewability.**
Permission check presence vs absence is visually identical in a diff. A reviewer looking at a new `CreateFoo` function has to remember to check that the block is there, that it's the right permission, and that the error maps correctly. The compiler can't enforce any of that. Over a team of 10 engineers and a year of work, something gets missed. The cost of that miss is a silent unauthorized-access bug.

## Proposed change — no-frame-works

Authorization lives on the **request struct** as a `CheckPermission` method, declared once per operation in the module's `api.go`. The method is called **exactly once**, at the top of every service function, in the same place it always is.

Permission *rules* (the boolean "can this actor do this thing?") live as pure methods on domain types. The request struct's `CheckPermission` is the dispatcher that pulls the actor from context, asks the domain, and translates a `false` into a sentinel error.

```go
// backend/internal/modules/post/api.go
func (r UpdatePostRequest) CheckPermission(ctx context.Context, post *domain.Post) error {
    a := actor.From(ctx)
    if a == nil {
        return ErrUnauthorized
    }
    if !post.CanModify(a) {
        return ErrForbidden
    }
    return nil
}
```

```go
// backend/internal/modules/post/domain/models.go
func (p Post) CanModify(a actor.Actor) bool {
    if a.IsSystem() {
        return true
    }
    if ua, ok := a.(actor.UserActor); ok && ua.HasRole(actor.RoleAdmin) {
        return true
    }
    return p.AuthorID == a.UserID().String()
}
```

Every service function that mutates a post calls this contract at the top, after `Validate`:

```go
// backend/internal/modules/post/service/update_post.go
func (s *Service) UpdatePost(ctx context.Context, req post.UpdatePostRequest) (*post.View, error) {
    if err := req.Validate(); err != nil {
        return nil, err
    }

    existing, err := s.repo.FindByID(ctx, req.ID)
    if err != nil {
        return nil, err
    }

    if permErr := req.CheckPermission(ctx, existing); permErr != nil {
        return nil, permErr
    }

    // ... business logic
}
```

Handlers do **not** call `CheckPermission`. Middleware does **not** call `CheckPermission`. The only place it is called is inside the service function, and it is always called, because the ritual is — `Validate` then `CheckPermission` then work, in that order, every time.

The handler's only job for authorization failures is to map the sentinel to an HTTP status code with `errors.Is`:

```go
if errors.Is(err, post.ErrForbidden) {
    return oapi.UpdatePost403JSONResponse{Error: "forbidden"}, nil
}
```

That mapping lives **once per sentinel per handler**, not once per endpoint.

## Why this is better

| Before (all three codebases) | After (no-frame-works) |
|------------------------------|------------------------|
| "Who can do what" is spread across 265 call sites | "Who can do what" is the set of `CheckPermission` methods in one `api.go` per module |
| Changing the permission system is a 258-file PR | Changing it is editing `CheckPermission` (one method) and the domain predicate (one method) |
| Forgetting to check permission is a silent 200 OK | Skipping `CheckPermission` is visible in review — the service ritual (Validate → CheckPermission → work) is a 3-line signature |
| Error → HTTP mapping duplicated at every call site | Error → HTTP mapping is one branch per sentinel per handler, using `errors.Is` |
| Permission params (`tenantId`, `orgId`, `userId`, `session`) threaded through every signature | Actor is on `context.Context`; `CheckPermission` extracts it once |
| Reviewing authorization correctness means reading the function body | Reviewing authorization correctness means checking `Validate()` and `CheckPermission()` appear at the top |
| Permission rule + dispatch + error mapping tangled at every call site | Rule (`CanModify` on domain), dispatch (`CheckPermission` on request), and mapping (`errors.Is` in handler) are three separate concerns |

The essential shift: **the request struct owns its authorization contract**. There is one place per operation where you can read "what this request needs," and the service function's job is to call that contract, not to re-implement it.

A second-order win: because permission rules are pure boolean methods on domain types (`post.CanModify(a actor.Actor) bool`), they are trivially unit-testable with a table-driven test. No mocks, no context, no DB. The domain becomes the single source of truth for authorization rules, and the request struct is the single source of truth for dispatch.

## Migration notes

For a codebase at `intranet-cms` scale, the migration is a per-service job but it's mechanical, and it doesn't need to happen all at once:

1. For each service function with an inline permission block, extract the request body into a new request struct inside the module's `api.go`.
2. Move the permission logic into a `CheckPermission` method on that struct. If the check requires the target entity (ownership), the method takes a second `*domain.X` parameter.
3. Move the boolean rule itself onto the domain type as a pure method (e.g. `CanModify`).
4. Replace the inline block in the service function with `req.CheckPermission(ctx, ...)`, placed immediately after `req.Validate()`.
5. Remove the naked `tenantId`, `orgId`, `userId` parameters from the service signature. They come from the request body and from the actor on context.
6. In the handler, replace the inline `errors.Is(err, useauth.ErrPermission)` branch with one that matches the new sentinel: `errors.Is(err, post.ErrForbidden)`.

No single PR can migrate 265 sites, and none should. The migration is driven by two rules:

- **Every new service function** uses the new pattern from day one. The count stops growing immediately.
- **Every time you touch an existing service function** for any reason, migrate it. The count drops monotonically as old code is touched.

The win is not "we fixed the whole codebase in a month." The win is that the problem stops compounding the moment the first PR lands, and that every subsequent change pays down a small amount of the debt automatically.
