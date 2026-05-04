# 03 — Repository Layer

> **Severity: HIGH**
> **Axes impacted: Maintainability • Comprehensibility • Reviewability**

## Responsibility

The repository is the project's **only gateway to the database**. Its job is:

1. Accept a **complete domain model** (for writes) or a simple identifier/filter (for reads).
2. Convert to the go-jet `model.*` type via a private `toModel` helper.
3. Execute a go-jet query — never raw SQL, never a different ORM.
4. Convert the result back via `toDomain` and return a `domain.*` type.

The repository does **not** know what "update a category" means as a business operation. It knows what "save a Category row" means as a database operation. The service decides when to save; the repository knows how.

Concretely:

- One operation per file, one exported function per operation.
- No partial-update methods (`UpdateName(id, name)` is wrong — `Update(ctx, c)` is right). The service sends a complete model; the repository persists it.
- The repository is accessed through a **module-owned `Repository` interface**, so services depend on an interface and implementations can be swapped or mocked.
- The "not found" contract is uniform: repositories return a domain sentinel (`domain.ErrXNotFound`) when a row does not exist. Never `(nil, nil)`.
- Timestamps and invariants belong in the domain factory, not in the repository. The repository persists whatever the domain says; the domain says when "now" is.

## What each repo does

### `intranet-cms` — flat `repository` package, timestamps in SQL, tagged errors

`backend/internal/app/repository/category.go:18-37`

```go
func CategoryTblInsert(ctx context.Context, db qrm.DB, m *model.CategoryTbl) (*model.CategoryTbl, error) {
    tbl := table.CategoryTbl
    m.CreatedAt = time.Now().UTC()
    m.UpdatedAt = time.Now().UTC()
    m.DeletedAt = nil

    smt := tbl.
        INSERT(tbl.MutableColumns).
        MODEL(m).
        RETURNING(tbl.AllColumns)
    result := []model.CategoryTbl{}
    err := smt.QueryContext(ctx, db, &result)
    if err != nil {
        return nil, fmt.Errorf("[#6rjv5uum] %w", err)
    }
    if len(result) != 1 {
        return nil, errors.New("[#79htf9z7]")
    }
    return &result[0], nil
}
```

Package-level function in a single flat `repository` package — no per-entity split. Read and write functions coexist: `CategoryTblInsert`, `CategoryTblUpdate`, `CategoryTblDelete` (soft delete), `CategoryTblBatchDelete`, `CategoryTblList`, `CategoryTblCountByOrganizationID`, and more. Takes `qrm.DB` as an executor parameter — the caller decides whether it is `app.DB()` or a `*sql.Tx`. Sets `CreatedAt`, `UpdatedAt`, `DeletedAt` inside the function. Errors are tagged with short codes (`[#6rjv5uum]`) but have no structured identity — every error is "something failed, here is a grep-able tag." Accepts and returns pointers to the go-jet `model.CategoryTbl` type.

### `workflows` — split `repository/` (reads) and `mutation/` (writes), multi-table writes in one mutation

`backend/internal/app/mutation/workflow_with_events_mutation.go:21-83`

```go
func InsertWorkflowWithEvents(
    ctx context.Context,
    db qrm.DB,
    workflowWithEvents domain_service.WorkflowWithEvents,
) (*WorkflowWithEventsMutationRes, error) {
    workflowWithEvents.Workflow.CreatedAt = workflowWithEvents.RequestTime()
    workflowWithEvents.Workflow.UpdatedAt = workflowWithEvents.RequestTime()
    workflowWithEvents.WorkflowVersion.CreatedAt = workflowWithEvents.RequestTime()
    workflowWithEvents.WorkflowVersion.UpdatedAt = workflowWithEvents.RequestTime()

    for i := range workflowWithEvents.NewEventLogs {
        // Increment request time by 1ms per event so they would be displayed in correct order
        offset := time.Duration(i) * time.Millisecond
        workflowWithEvents.NewEventLogs[i].CreatedAt = workflowWithEvents.RequestTime().Add(offset)
        workflowWithEvents.NewEventLogs[i].UpdatedAt = workflowWithEvents.RequestTime().Add(offset)
    }

    // Insert workflow
    insertWorkflowStmt := workflowTbl.INSERT(workflowTbl.AllColumns).MODEL(workflowWithEvents.Workflow).RETURNING(workflowTbl.AllColumns)
    err := insertWorkflowStmt.QueryContext(ctx, db, &newWorkflows)
    if err != nil || len(newWorkflows) != 1 { return nil, err }

    // Insert workflow version
    insertWorkflowVersionStmt := workflowVersionTbl.INSERT(workflowVersionTbl.AllColumns).MODEL(workflowWithEvents.WorkflowVersion).RETURNING(workflowVersionTbl.AllColumns)
    err = insertWorkflowVersionStmt.QueryContext(ctx, db, &newWorkflowVersions)
    if err != nil || len(newWorkflowVersions) != 1 { return nil, err }

    // Insert workflow event logs
    insertEventLogStmt := workflowEventLogTbl.INSERT(workflowEventLogTbl.MutableColumns).MODELS(workflowWithEvents.NewEventLogs).RETURNING(workflowEventLogTbl.AllColumns)
    err = insertEventLogStmt.QueryContext(ctx, db, &newEventLogs)
    if err != nil || len(newEventLogs) != len(workflowWithEvents.NewEventLogs) { return nil, err }

    return &WorkflowWithEventsMutationRes{Workflow: newWorkflows[0], WorkflowVersion: newWorkflowVersions[0], EventLogs: newEventLogs}, nil
}
```

Split-package layout: reads in `repository/`, writes in `mutation/`. Functions are package-level, take `qrm.DB`, and accept types owned by other packages (`domain_service.WorkflowWithEvents`). This single mutation writes to **three tables** in sequence: workflow, workflow version, event logs — so the mutation contains orchestration logic that should live in the service. It also contains business logic inline: event log timestamps are offset by `i * time.Millisecond` to preserve display order. Error handling is `return nil, err` with no wrapping — the caller has no context for which statement failed.

### `id-services v3` — split `repository/` + `mutation/`, inconsistent "not found" contract

`backend/internal/app/repository/organization_repo/repo.go:129-152`

```go
func GetByTenantAndName(
    ctx context.Context,
    db qrm.DB,
    tenantId uuid.UUID,
    name string,
) (*model.Organization, error) {
    tbl := table.Organization
    stmt := pg.SELECT(tbl.AllColumns).
        FROM(tbl).
        WHERE(tbl.TenantID.EQ(pg.UUID(tenantId)).AND(tbl.Name.EQ(pg.String(name))))

    rows := []model.Organization{}
    err := stmt.QueryContext(ctx, db, &rows)
    if err != nil {
        return nil, errors.Wrap(err, "failed to execute SQL query")
    }

    if len(rows) == 0 {
        return nil, nil  // ← not an error, just nil pointer
    }

    return &rows[0], nil
}
```

`backend/internal/app/mutation/organization.go:14-33`

```go
func InsertOrganization(
    ctx context.Context,
    db qrm.DB,
    row model.Organization,
) (*model.Organization, error) {
    now := time.Now().UTC()
    row.CreatedAt = &now
    row.UpdatedAt = &now

    tbl := table.Organization
    stmt := tbl.INSERT(tbl.AllColumns).MODEL(row).RETURNING(tbl.AllColumns)

    rows := []model.Organization{}
    err := stmt.QueryContext(ctx, db, &rows)
    if err != nil {
        return nil, utils.WrapErr(err)
    }

    return &rows[0], nil
}
```

Split-package layout similar to `workflows`. Reads live in `<entity>_repo/` packages, writes in `mutation/`. One function per file; every function has a matching `_test.go`. Errors are wrapped with `errors.Wrap(err, "...")` or `utils.WrapErr(err)` for structured stack traces.

But the **"not found" contract is inconsistent**. `GetByTenantAndName` returns `(nil, nil)` when the row does not exist. The sibling function in the same file, `GetByTenantAndId` (line 154), returns `errors.ErrDbEmptyResult` for the same case. Callers have to know per-function which convention applies — and nothing in the type signature tells them. This is the kind of inconsistency that produces nil-pointer bugs on the third Tuesday of the quarter.

## What's good

- **All three use go-jet.** No raw SQL, type-safe column references.
- **`workflows` and `id-services v3` split reads and writes into separate packages.** Stronger auditability than `intranet-cms`'s flat `repository` — "who mutates what" is obvious from the import path.
- **All three thread a `qrm.DB` executor through every function.** This is exactly what `shared.GetExecutor(ctx, r.db)` does in the `no-frame-works` template — the repository is transaction-agnostic because the caller hands in the executor. The intent is right; only the mechanism is crude.
- **`id-services v3` has one function per file with tests** (`create.go` + `create_test.go`, `get_by_id.go` + `get_by_id_test.go`). Same layout the `no-frame-works` template recommends.
- **`intranet-cms` uses `INSERT(MutableColumns)`** for inserts — go-jet's idiomatic way to skip auto-generated columns.
- **`id-services v3` and `workflows` wrap errors with a package helper** (`errors.Wrap`, `utils.WrapErr`), which carries stack traces into logs.

## What's bad

- **No interface anywhere.** None of the three codebases has a `Repository` interface that services depend on. Services call concrete package functions. This means:
  - You cannot mock the repository in a service test — tests hit a real database, or do not test the service layer at all.
  - You cannot swap the database or the ORM without rewriting every service.
  - The "service depends on repo" arrow is implicit in imports, not explicit in a type.
- **`model.*` types cross the service boundary.** In all three, the repository returns go-jet-generated `model.*` structs (or, in `workflows`, internal DTOs aliased to oapi types). The service then treats these as its own domain objects. There is no `domain.*` layer between storage and business logic.
- **Timestamp concerns live in the repository.** `intranet-cms` sets `CreatedAt`/`UpdatedAt`/`DeletedAt` inside `CategoryTblInsert`. `workflows` sets workflow + version timestamps and *also* loops to offset event-log timestamps for ordering inside `InsertWorkflowWithEvents` (business logic masquerading as data access). `id-services v3` sets timestamps inside `InsertOrganization`. "When was this created" is a domain fact, not a data-access concern — it belongs in the domain factory.
- **Read/write split is a poor substitute for layered architecture.** Splitting `repository/` from `mutation/` gives you auditability (a `grep mutation.` audit reveals every write) but does nothing for the deeper problem: the call shape and types are unchanged. A service still imports from both packages and still sees `model.*` types.
- **`workflows` mutations do multi-table writes in one function.** `InsertWorkflowWithEvents` writes to three tables in sequence. That is fine for atomicity (a transaction handles rollback) but bad for repository simplicity — the mutation is doing orchestration, which is service work. Each table should have its own insert function; the service should call them inside a transaction.
- **Inconsistent "not found" contract in `id-services v3`.** Some functions return `(nil, nil)` for missing rows; others return `errors.ErrDbEmptyResult`. Callers cannot rely on a uniform handling pattern.
- **`workflows` returns bare `err` without wrapping.** `return nil, err` with no context — the caller has no idea which of the three INSERTs failed.
- **Error handling in `intranet-cms` is tagged but not structured.** `fmt.Errorf("[#6rjv5uum] %w", err)` carries a grep-able tag but no sentinel identity. A handler cannot use `errors.Is` to distinguish "row already exists" from "connection lost."

## Improvements — one repo at a time

### `intranet-cms`

1. **Split the `repository` package per-entity.** `repository/category/`, `repository/user_group/`, etc. Within each, one file per operation (`create.go`, `find_by_id.go`, `update.go`, ...).
2. **Declare a `Repository` interface in the module package** (`category.Repository`). The service depends only on the interface; tests use a fake.
3. **Introduce `domain.*` types and private `toModel` / `toDomain` helpers in the repository.** Stop returning `*model.CategoryTbl` from the repository — return `*domain.Category`.
4. **Move timestamp setting into the domain factory.** `domain.NewCategory(...)` sets `CreatedAt`/`UpdatedAt` at construction. The repository persists whatever the domain says.
5. **Return domain sentinels for structured outcomes.** `domain.ErrCategoryNotFound` on zero rows; the handler maps sentinels to HTTP status via `errors.Is`.
6. **Replace the per-function `qrm.DB` parameter with `shared.GetExecutor(ctx, r.db)`** inside the repository. Services stop passing executors around.

### `workflows`

1. **Collapse `repository/` + `mutation/` into per-entity packages** with a read+write `Repository` interface per module. The module-scoped interface gives you the same auditability as the read/write split (grep for `s.repo.*`) without the package fragmentation.
2. **Define `domain.*` types to replace `internal_dto.*`.** The current internal DTO layer is a type alias over oapi — there is nothing truly domain-owned. Build real domain types with factories and methods.
3. **Break `InsertWorkflowWithEvents` into three repository methods** (`workflowRepo.Create`, `workflowVersionRepo.Create`, `eventLogRepo.InsertBatch`) and orchestrate them in the service inside a `TxManager.WithTransaction` block. The mutation stops being a mini-service.
4. **Move event log timestamp offset logic out of the mutation.** That is business logic ("preserve display order"); it belongs on `domain.EventLog` or on the service that constructs the event log list.
5. **Wrap errors with statement context.** `return nil, errors.Errorf("insert workflow: %w", err)` instead of `return nil, err`. The caller reads the log and knows which INSERT failed.
6. **Return domain sentinels for structured outcomes** (`domain.ErrWorkflowTemplateNotFound`, `domain.ErrVersionConflict`).

### `id-services v3`

1. **Declare a `team.Repository` interface** in the module. Move `organization_repo` and `mutation` functions behind implementations of it. Services depend on the interface, not the concrete packages.
2. **Introduce `domain.Team`** distinct from `model.Organization`. Repository takes `domain.Team` in, returns `*domain.Team` out. `toModel` / `toDomain` are private to the repository package.
3. **Unify the "not found" contract.** Every `FindBy*` function returns `domain.ErrTeamNotFound` on zero rows. No `(nil, nil)` — it is ambiguous and invites nil-pointer bugs in the caller.
4. **Move timestamp setting into the domain factory.** `domain.NewTeam(...)` sets `CreatedAt`/`UpdatedAt`; `InsertOrganization`'s job shrinks to "execute the INSERT."
5. **Adopt a `TxManager` with context-based propagation.** Right now `teams.Create` has no transaction wrapping because it does exactly one INSERT. The moment a second write is added (e.g. creating a default group alongside a team), you need atomicity — and the current shape has no clean way to declare it.
6. **Keep the one-file-per-function layout and test coverage.** This is the best thing v3 already does; the improvements above should preserve it, not replace it.
