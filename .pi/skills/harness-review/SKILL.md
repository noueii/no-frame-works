---
name: harness-review
description: >
  Reviews Go backend code for architecture violations against the no-frame-works
  layered architecture rules. Use /harness-review when you need a full review of
  code changes, new endpoints, or to verify compliance with handler/service/repository/domain
  layer separation. Checks type isolation, error handling, and request flow integrity.
---

# Harness Review Skill

Run a comprehensive review of Go backend code for architecture violations.

## Setup

None required. Rules are in `.agents/rules/` in the project.

## Usage

```
/harness-review [target]
```

Where `target` is optional — file, directory, or files pattern to review.

Examples:
- `/harness-review` — review all modified/new files
- `/harness-review handler/post_create_post.go` — review specific handler
- `/harness-review backend/internal/app/services/` — review entire services folder

## Review Rubrics

Read the relevant rubric files from `.agents/rules/`:

1. **Handler Layer** — `.agents/rules/handler.md`
   - Only oapi types + API contract types
   - No domain.* imports
   - Pure transformers, no business logic

2. **Service Layer** — `.agents/rules/service.md`
   - Calls `req.Validate()` first, then `req.CheckPermission()`
   - Returns view types, not domain models
   - One function per file

3. **Repository Layer** — `.agents/rules/repository.md`
   - No raw SQL — go-jet only
   - Uses `MODEL()` with `MutableColumns` for updates
   - `toModel`/`toDomain` mapping functions here

4. **Domain Layer** — `.agents/rules/domain.md`
   - No infrastructure imports (database/sql, net/http, external SDKs)
   - Pure business logic only
   - Sentinel errors in `domain/errors.go`

5. **Flow Integrity** — `.agents/rules/flow.md`
   - Every step in the request lifecycle present
   - Correct types at each boundary
   - No layer bypassing

6. **Common Mistakes** — `.agents/rules/patterns.md`
   - `fmt.Errorf` vs `errors.Errorf`
   - stdlib errors vs go-errors
   - %w wrapping

7. **Caveman Review** — `.agents/caveman-review.md`
   - Terse, actionable comments
   - Severity prefix (🔴🟠🟡🔵❓)
   - Location, problem, fix format

## Review Steps

1. **Identify changed/new files** — Use git status or scan target directory
2. **Classify by layer** — handler, service, repository, domain, or cross-layer
3. **Apply relevant rubric** — Read the rubric file, then inspect the code
4. **Check error patterns** — Search for `fmt.Errorf`, stdlib `errors`, raw SQL
5. **Trace the flow** — For new endpoints, verify all steps exist
6. **Output in caveman-review format** — One line per finding

## Output Format

Use caveman-review format for findings:

```
🔴 bug: <file>:L<line> — <problem>. <fix>.
🟠 conv: <file>:L<line> — <violation>. <correct pattern>.
🟡 risk: <file>:L<line> — <fragile pattern>. <robust alternative>.
🔵 nit: <file>:L<line> — <style issue>.
❓ q: <file>:L<line> — <genuine question>.
```

Start with a brief verdict (good/needs work), then list findings.

## Auto-Trigger

If you see code that clearly violates the architecture while working:
- Flag it immediately in your response
- Explain why it's wrong
- Show the correct pattern
- Don't wait for explicit review request