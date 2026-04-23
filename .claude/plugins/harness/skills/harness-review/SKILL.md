---
name: harness-review
description: >
  Reviews Go backend code for violations of the no-frame-works layered architecture.
  Activates when asked to review code, check architecture, or verify compliance with
  handler/service/repository/domain layer rules. Also use after implementing new endpoints.
---

# Go Backend Harness Review

You are reviewing Go backend code against the no-frame-works layered architecture rules.

## Rules Location

Project rules are in `.agents/rules/`:
- `handler.md` — handler layer rules
- `service.md` — service layer rules  
- `repository.md` — repository layer rules
- `domain.md` — domain layer rules
- `flow.md` — request flow integrity
- `patterns.md` — common mistakes

## Architecture Overview

### Type Isolation
| Layer | May use | Must NOT use |
|-------|---------|--------------|
| Handler | `oapi.*`, API contract types | `domain.*`, `model.*` |
| Service | API contract, `domain.*` | `oapi.*`, `model.*` |
| Repository | `domain.*`, `model.*` | `oapi.*`, API contract |
| Domain | pure types only | infrastructure (`database/sql`, `net/http`) |

### Request Flow
```
oapi request → handler transforms → service request → req.Validate() → 
req.CheckPermission() → domain model → repo (toModel) → go-jet query → 
repo (toDomain) → view → oapi response
```

### Error Rules
- Use `github.com/go-errors/errors` — NEVER `fmt.Errorf` or stdlib `errors`
- Wrap with `%w`: `errors.Errorf("layer.service.op: %w", err)`
- Six shared sentinels: `ErrNotFound`, `ErrValidation`, `ErrUnauthorized`, `ErrForbidden`, `ErrConflict`, `ErrInternal`
- Log once at handler, never silently swallow errors

## Review Process

1. **Identify changed files** — check git changes or scan target directory
2. **Classify by layer** — handler, service, repository, domain, or cross-layer
3. **Read relevant rubric** — use read tool to load the rule file
4. **Apply checks**:
   - Type imports at each boundary
   - Validate()/CheckPermission() in services
   - go-jet usage (no raw SQL) in repos
   - Pure domain (no infrastructure imports)
   - Error patterns (fmt.Errorf vs errors.Errorf)
5. **Output findings** in caveman-review format

## Output Format

Use caveman-review format with severity prefix:

```
🔴 bug: <file>:L<line> — <problem>. <fix>.
🟠 conv: <file>:L<line> — <violation>. <correct pattern>.
🟡 risk: <file>:L<line> — <fragile pattern>. <robust alternative>.
🔵 nit: <file>:L<line> — <style issue>.
❓ q: <file>:L<line> — <question for author>.
```

Start with overall verdict, then detailed findings grouped by layer.

## Severity Guide

| Prefix | Meaning | When to use |
|--------|---------|-------------|
| 🔴 bug | Will cause incident | Broken behavior, crashes, data loss |
| 🟠 conv | Architecture violation | Wrong layer, wrong type, skipped step |
| 🟡 risk | Works but fragile | Race conditions, missing null check |
| 🔵 nit | Style | Naming, formatting, minor issues |
| ❓ q | Question | Need clarification from author |

## Auto-Trigger

If you see clear architecture violations while working:
- Flag them immediately
- Explain why it's wrong
- Show the correct pattern
- Don't wait for explicit review request